package txn

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/logger"
	"golang.org/x/sync/errgroup"
)

type GroupKeyMaintainer struct {
	connMap map[string]Connector
}

func NewGroupKeyMaintainer() *GroupKeyMaintainer {
	return &GroupKeyMaintainer{
		connMap: make(map[string]Connector),
	}
}

func getList(item DataItem) []string {
	listStr := item.GroupKeyList()
	// group keys are split by whitespace
	list := strings.Split(listStr, " ")
	return list
}

func (g *GroupKeyMaintainer) AddConnector(ds Datastorer) {
	g.connMap[ds.GetName()] = ds.GetConn()
}

func (g *GroupKeyMaintainer) GetGroupKey(urls []string) ([]GroupKey, error) {
	groupKeys := make([]GroupKey, 0, len(urls))
	var mu sync.Mutex
	var eg errgroup.Group
	for _, urll := range urls {
		url := urll
		eg.Go(func() error {
			groupKey, err := g.GetSingleGroupKey(url)
			if err != nil {
				return err
			}
			mu.Lock()
			groupKeys = append(groupKeys, groupKey)
			mu.Unlock()
			return nil
		})
	}
	err := eg.Wait()
	// This means at least one of the group key is not found
	if err != nil {
		return nil, err
	}

	return groupKeys, nil
}

func (g *GroupKeyMaintainer) GetSingleGroupKey(url string) (GroupKey, error) {
	tokens := strings.Split(url, ":")
	conn, ok := g.connMap[tokens[0]]
	if !ok {
		return GroupKey{}, fmt.Errorf("Connector to %s is not found", tokens[0])
	}
	groupKeyStr, err := conn.Get(url)
	if err != nil {
		return GroupKey{}, err
	}
	var keyItem GroupKeyItem
	err = json.Unmarshal([]byte(groupKeyStr), &keyItem)
	if err != nil {
		return GroupKey{}, fmt.Errorf("failed to unmarshal group key item %s", groupKeyStr)
	}
	return *NewGroupKey(url, keyItem.TxnState, keyItem.TCommit), nil
}

// GetGroupKeyList reads a group key list for a transaction.
// It will return an error if at least one of the group key is not found.
func (g *GroupKeyMaintainer) GetGroupKeyList(item DataItem) ([]GroupKey, error) {
	keyList := getList(item)
	return g.GetGroupKey(keyList)
}

func (g *GroupKeyMaintainer) CreateGroupKey(urls []string, state config.State) int {
	resChan := make(chan error, len(urls))
	for _, urll := range urls {
		url := urll
		go func() {
			tokens := strings.Split(url, ":")
			conn, ok := g.connMap[tokens[0]]
			if !ok {
				resChan <- fmt.Errorf("Connector to %s is not found", tokens[0])
				return
			}

			groupKey := NewGroupKey(url, state, 0)
			groupKeyStr, err := json.Marshal(groupKey)
			if err != nil {
				resChan <- fmt.Errorf("failed to marshal group key item %s", groupKey)
				return
			}
			// CHECK: we do not need the returned value?
			_, err = conn.AtomicCreate(url, util.ToString(groupKeyStr))
			if err != nil {
				resChan <- err
				return
			}
			resChan <- nil
		}()
	}
	okInTotal := len(urls)
	for i := 0; i < len(urls); i++ {
		err := <-resChan
		if err != nil {
			logger.Log.Errorw("Error in creating group key", "error", err)
			okInTotal--
		}
	}
	return okInTotal
}

// CreateGroupKeyList creates a group key list for a transaction.
// returns the number of group keys successfully created.
func (g *GroupKeyMaintainer) CreateGroupKeyList(item DataItem, state config.State) int {
	keyList := getList(item)
	resChan := make(chan error, len(keyList))
	for _, urll := range keyList {
		url := urll
		go func() {
			tokens := strings.Split(url, ":")
			conn, ok := g.connMap[tokens[0]]
			if !ok {
				resChan <- fmt.Errorf("Connector to %s is not found; url: %s, item: %v", tokens[0], url, item)
				return
			}

			groupKey := NewGroupKey(url, state, 0)
			groupKeyStr, err := json.Marshal(groupKey)
			if err != nil {
				resChan <- fmt.Errorf("failed to marshal group key item %s", groupKey)
				return
			}
			// CHECK: we do not need the returned value?
			_, err = conn.AtomicCreate(url, util.ToString(groupKeyStr))
			if err != nil {
				resChan <- err
				return
			}
			resChan <- nil
		}()
	}
	okInTotal := len(keyList)
	for i := 0; i < len(keyList); i++ {
		err := <-resChan
		if err != nil {
			logger.Log.Errorw("Error in creating group key", "error", err)
			okInTotal--
		}
	}
	return okInTotal
}

func (g *GroupKeyMaintainer) DeleteGroupKeyList(item DataItem) error {
	keyList := getList(item)
	return g.DeleteGroupKey(keyList)
}

func (g *GroupKeyMaintainer) DeleteGroupKey(urls []string) error {
	var eg errgroup.Group
	for _, urll := range urls {
		url := urll
		eg.Go(func() error {
			tokens := strings.Split(url, ":")
			conn, ok := g.connMap[tokens[0]]
			if !ok {
				return fmt.Errorf("Connector to %s is not found", tokens[0])
			}
			err := conn.Delete(url)
			return err
		})
	}
	return eg.Wait()
}
