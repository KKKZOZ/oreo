package workload

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"benchmark/pkg/benconfig"
	"benchmark/pkg/generator"
	"benchmark/pkg/util"
	"benchmark/ycsb"
)

// Data structures modeling DeathStarBench Social Network

type User struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	Bio         string    `json:"bio"`
	AvatarURL   string    `json:"avatar_url"`
	CreatedAt   time.Time `json:"created_at"`
	LastLoginAt time.Time `json:"last_login_at"`
}

type Post struct {
	PostID       string    `json:"post_id"`
	UserID       string    `json:"user_id"`
	Text         string    `json:"text"`
	Mentions     []string  `json:"mentions"`   // user mentions
	MediaURLs    []string  `json:"media_urls"` // media attachments
	URLs         []string  `json:"urls"`       // shortened URLs
	PostType     string    `json:"post_type"`  // post, repost, reply
	CreatedAt    time.Time `json:"created_at"`
	LikeCount    int       `json:"like_count"`
	CommentCount int       `json:"comment_count"`
}

type Comment struct {
	CommentID string    `json:"comment_id"`
	PostID    string    `json:"post_id"`
	UserID    string    `json:"user_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type TimelineEntry struct {
	PostID    string    `json:"post_id"`
	UserID    string    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

type SocialGraph struct {
	UserID         string   `json:"user_id"`
	Followers      []string `json:"followers"`
	Following      []string `json:"following"`
	FollowerCount  int      `json:"follower_count"`
	FollowingCount int      `json:"following_count"`
}

type UserSession struct {
	UserID     string    `json:"user_id"`
	SessionID  string    `json:"session_id"`
	LastAccess time.Time `json:"last_access"`
	IPAddress  string    `json:"ip_address"`
}

type Media struct {
	MediaID   string    `json:"media_id"`
	UserID    string    `json:"user_id"`
	URL       string    `json:"url"`
	MediaType string    `json:"media_type"` // image, video, gif
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

type SocialNetworkWorkload struct {
	mu sync.Mutex

	Randomizer
	taskChooser *generator.Discrete
	wp          *WorkloadParameter

	// DeathStarBench-style service namespaces
	UserServiceNS         string // user management
	SocialGraphServiceNS  string // follower/following relationships
	PostServiceNS         string // post storage and retrieval
	TimelineServiceNS     string // timeline generation
	ComposeServiceNS      string // post composition
	MediaServiceNS        string // media handling
	TextServiceNS         string // text processing
	URLServiceNS          string // URL shortening
	UniqueIDServiceNS     string // unique ID generation
	HomeTimelineServiceNS string // home timeline cache
	UserTimelineServiceNS string // user timeline

	// Task counters
	task1Count  int // ComposePost
	task2Count  int // ReadHomeTimeline
	task3Count  int // ReadUserTimeline
	task4Count  int // FollowUser
	task5Count  int // UnfollowUser
	task6Count  int // LikePost
	task7Count  int // CommentOnPost
	task8Count  int // GetUserProfile
	task9Count  int // SearchUsers
	task10Count int // GetRecommendedUsers

	// Sample data
	usernames  []string
	firstNames []string
	lastNames  []string
	postTexts  []string
	mediaTypes []string
}

var _ Workload = (*SocialNetworkWorkload)(nil)

func NewSocialNetworkWorkload(wp *WorkloadParameter) *SocialNetworkWorkload {
	return &SocialNetworkWorkload{
		mu:                    sync.Mutex{},
		Randomizer:            *NewRandomizer(wp),
		taskChooser:           createTaskGenerator(wp),
		wp:                    wp,
		UserServiceNS:         "user",
		SocialGraphServiceNS:  "social-graph",
		PostServiceNS:         "post",
		TimelineServiceNS:     "timeline",
		ComposeServiceNS:      "compose",
		MediaServiceNS:        "media",
		TextServiceNS:         "text",
		URLServiceNS:          "url-shorten",
		UniqueIDServiceNS:     "unique-id",
		HomeTimelineServiceNS: "home-timeline",
		UserTimelineServiceNS: "user-timeline",
		usernames: []string{
			"john_doe", "jane_smith", "tech_guru", "foodie_lover", "travel_bug",
			"music_fan", "bookworm", "fitness_pro", "artist_soul", "code_ninja",
			"photo_enthusiast", "gamer_elite", "news_junkie", "movie_buff", "sports_fan",
		},
		firstNames: []string{
			"John", "Jane", "Mike", "Sarah", "David",
			"Lisa", "Tom", "Emily", "Chris", "Anna",
			"James", "Emma", "Robert", "Olivia", "William",
		},
		lastNames: []string{
			"Smith", "Johnson", "Williams", "Brown", "Jones",
			"Garcia", "Miller", "Davis", "Rodriguez", "Martinez",
			"Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson",
		},
		postTexts: []string{
			"Just had an amazing day!",
			"Check out this cool new feature!",
			"What a beautiful sunset today",
			"Learning something new every day",
			"Can't wait for the weekend!",
			"This is so inspiring!",
			"Just finished reading an amazing book",
			"Technology is changing so fast",
			"Great workout session today",
			"Delicious meal at the new restaurant",
		},
		mediaTypes: []string{"image", "video", "gif"},
	}
}

func (wl *SocialNetworkWorkload) generateUser(userID string) User {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	firstName := wl.firstNames[wl.r.Intn(len(wl.firstNames))]
	lastName := wl.lastNames[wl.r.Intn(len(wl.lastNames))]
	username := wl.usernames[wl.r.Intn(len(wl.usernames))] + "_" + userID

	return User{
		UserID:      userID,
		Username:    username,
		FirstName:   firstName,
		LastName:    lastName,
		Email:       fmt.Sprintf("%s.%s@example.com", firstName, lastName),
		Bio:         fmt.Sprintf("Hello, I'm %s %s", firstName, lastName),
		AvatarURL:   fmt.Sprintf("https://cdn.social.com/avatars/%s.jpg", userID),
		CreatedAt:   time.Now(),
		LastLoginAt: time.Now(),
	}
}

func (wl *SocialNetworkWorkload) generatePost(postID, userID string) Post {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	text := wl.postTexts[wl.r.Intn(len(wl.postTexts))]

	// Add random mentions (0-3)
	mentionCount := wl.r.Intn(4)
	mentions := make([]string, mentionCount)
	for i := 0; i < mentionCount; i++ {
		mentions[i] = fmt.Sprintf("user%d", wl.r.Intn(10000))
	}

	// Add random media (0-4)
	mediaCount := wl.r.Intn(5)
	mediaURLs := make([]string, mediaCount)
	for i := 0; i < mediaCount; i++ {
		mediaURLs[i] = fmt.Sprintf("https://cdn.social.com/media/%s_%d.jpg", postID, i)
	}

	return Post{
		PostID:       postID,
		UserID:       userID,
		Text:         text,
		Mentions:     mentions,
		MediaURLs:    mediaURLs,
		URLs:         []string{},
		PostType:     "post",
		CreatedAt:    time.Now(),
		LikeCount:    0,
		CommentCount: 0,
	}
}

func (wl *SocialNetworkWorkload) generateComment(commentID, postID, userID string) Comment {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	comments := []string{
		"Great post!",
		"I totally agree!",
		"Thanks for sharing!",
		"This is amazing!",
		"Love this!",
		"Interesting perspective",
	}

	return Comment{
		CommentID: commentID,
		PostID:    postID,
		UserID:    userID,
		Text:      comments[wl.r.Intn(len(comments))],
		CreatedAt: time.Now(),
	}
}

func (wl *SocialNetworkWorkload) generateMedia(mediaID, userID string) Media {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	mediaType := wl.mediaTypes[wl.r.Intn(len(wl.mediaTypes))]
	size := int64(wl.r.Intn(10000000) + 100000) // 100KB - 10MB

	return Media{
		MediaID:   mediaID,
		UserID:    userID,
		URL:       fmt.Sprintf("https://cdn.social.com/media/%s.%s", mediaID, mediaType),
		MediaType: mediaType,
		Size:      size,
		CreatedAt: time.Now(),
	}
}

// Task 1: ComposePost - DeathStarBench: Complete post composition workflow
func (wl *SocialNetworkWorkload) ComposePost(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()

	userID := wl.NextKeyName()
	postID := wl.NextKeyName()

	// 1. Verify user authentication (User Service)
	_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.UserServiceNS, userID))

	// 2. Generate unique post ID (Unique ID Service) - stored in KVRocks
	db.Update(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:counter", wl.UniqueIDServiceNS),
		postID,
	)

	// 3. Process text and extract URLs/mentions (Text Service)
	post := wl.generatePost(postID, userID)
	textProcessing := map[string]interface{}{
		"original_text": post.Text,
		"mentions":      post.Mentions,
		"urls":          post.URLs,
	}
	textData, _ := json.Marshal(textProcessing)
	db.Update(
		ctx,
		"Redis",
		fmt.Sprintf("%v:process:%v", wl.TextServiceNS, postID),
		string(textData),
	)

	// 4. Store media if present (Media Service) - metadata in KVRocks
	if len(post.MediaURLs) > 0 {
		for i := range post.MediaURLs {
			mediaID := fmt.Sprintf("%s_media_%d", postID, i)
			media := wl.generateMedia(mediaID, userID)
			mediaData, _ := json.Marshal(media)
			db.Update(
				ctx,
				"KVRocks",
				fmt.Sprintf("%v:%v", wl.MediaServiceNS, mediaID),
				string(mediaData),
			)
		}
	}

	// 5. Store post (Post Service) - in MongoDB
	postData, _ := json.Marshal(post)
	db.Update(
		ctx,
		"MongoDB2",
		fmt.Sprintf("%v:%v", wl.PostServiceNS, postID),
		string(postData),
	)

	// 6. Add to user's timeline (User Timeline Service) - in Cassandra
	timelineEntry := TimelineEntry{
		PostID:    postID,
		UserID:    userID,
		Timestamp: time.Now(),
	}
	timelineData, _ := json.Marshal(timelineEntry)
	db.Update(
		ctx,
		"Cassandra",
		fmt.Sprintf("%v:%v:%v", wl.UserTimelineServiceNS, userID, postID),
		string(timelineData),
	)

	// 7. Fan-out to followers' home timelines (Home Timeline Service) - cached in Redis
	// Simulate fan-out to 5 followers
	for i := 0; i < 5; i++ {
		followerID := wl.NextKeyName()
		db.Update(
			ctx,
			"Redis",
			fmt.Sprintf("%v:%v:%v", wl.HomeTimelineServiceNS, followerID, postID),
			string(timelineData),
		)
	}

	// 8. Update user's post count (Social Graph Service) - in KVRocks
	db.Update(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:%v:post_count", wl.SocialGraphServiceNS, userID),
		"1",
	)

	db.Commit()
}

// Task 2: ReadHomeTimeline - DeathStarBench: Read user's home timeline
func (wl *SocialNetworkWorkload) ReadHomeTimeline(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()

	userID := wl.NextKeyName()

	// 1. Check user session (Redis cache)
	_, _ = db.Read(
		ctx,
		"Redis",
		fmt.Sprintf("%v:session:%v", wl.UserServiceNS, userID),
	)

	// 2. Try to read from Redis cache first (Home Timeline Service)
	_, _ = db.Read(
		ctx,
		"Redis",
		fmt.Sprintf("%v:%v:cache", wl.HomeTimelineServiceNS, userID),
	)

	// 3. If cache miss, read timeline entries (simulate reading 3 posts)
	for i := 0; i < 3; i++ {
		postID := wl.NextKeyName()

		// Read from Redis hot cache
		_, _ = db.Read(
			ctx,
			"Redis",
			fmt.Sprintf("%v:%v:%v", wl.HomeTimelineServiceNS, userID, postID),
		)

		// Read post content (Post Service)
		_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.PostServiceNS, postID))

		// Read post author info (User Service)
		authorID := wl.NextKeyName()
		_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.UserServiceNS, authorID))
	}

	// 4. Cache the assembled timeline in Redis
	timelineCache := map[string]interface{}{
		"user_id":    userID,
		"post_count": 10,
		"cached_at":  time.Now(),
	}
	cacheData, _ := json.Marshal(timelineCache)
	db.Update(
		ctx,
		"Redis",
		fmt.Sprintf("%v:%v:cache", wl.HomeTimelineServiceNS, userID),
		string(cacheData),
	)

	db.Commit()
}

// Task 3: ReadUserTimeline - DeathStarBench: Read specific user's timeline
func (wl *SocialNetworkWorkload) ReadUserTimeline(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()

	userID := wl.NextKeyName()
	viewerID := wl.NextKeyName()

	// 1. Get user profile (User Service)
	_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.UserServiceNS, userID))

	// 2. Get user's post count from social graph (KVRocks)
	_, _ = db.Read(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:%v:post_count", wl.SocialGraphServiceNS, userID),
	)

	// 3. Read user timeline from Cassandra (time-ordered)
	for i := 0; i < 3; i++ {
		postID := wl.NextKeyName()
		_, _ = db.Read(
			ctx,
			"Cassandra",
			fmt.Sprintf("%v:%v:%v", wl.UserTimelineServiceNS, userID, postID),
		)

		// Get post details
		_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.PostServiceNS, postID))

		// Get like count from Redis
		_, _ = db.Read(
			ctx,
			"Redis",
			fmt.Sprintf("%v:%v:likes", wl.PostServiceNS, postID),
		)
	}

	// 4. Update viewer session
	session := UserSession{
		UserID:     viewerID,
		SessionID:  fmt.Sprintf("session_%d", wl.r.Intn(999999)),
		LastAccess: time.Now(),
		IPAddress:  fmt.Sprintf("192.168.%d.%d", wl.r.Intn(256), wl.r.Intn(256)),
	}
	sessionData, _ := json.Marshal(session)
	db.Update(
		ctx,
		"Redis",
		fmt.Sprintf("%v:session:%v", wl.UserServiceNS, viewerID),
		string(sessionData),
	)

	db.Commit()
}

// Task 4: FollowUser - DeathStarBench: Follow a user
func (wl *SocialNetworkWorkload) FollowUser(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()

	followerID := wl.NextKeyName()
	followeeID := wl.NextKeyName()

	// 1. Verify both users exist (User Service)
	_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.UserServiceNS, followerID))
	_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.UserServiceNS, followeeID))

	// 2. Update follower's following list (Social Graph Service) - in KVRocks
	followData := map[string]interface{}{
		"follower_id": followerID,
		"followee_id": followeeID,
		"created_at":  time.Now(),
	}
	followDataJSON, _ := json.Marshal(followData)
	db.Update(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:%v:following:%v", wl.SocialGraphServiceNS, followerID, followeeID),
		string(followDataJSON),
	)

	// 3. Update followee's followers list (Social Graph Service) - in KVRocks
	db.Update(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:%v:followers:%v", wl.SocialGraphServiceNS, followeeID, followerID),
		string(followDataJSON),
	)

	// 4. Update follower count (KVRocks)
	db.Update(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:%v:follower_count", wl.SocialGraphServiceNS, followeeID),
		"1",
	)

	// 5. Update following count (KVRocks)
	db.Update(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:%v:following_count", wl.SocialGraphServiceNS, followerID),
		"1",
	)

	// 6. Invalidate home timeline cache (Redis)
	db.Update(
		ctx,
		"Redis",
		fmt.Sprintf("%v:%v:cache:invalidate", wl.HomeTimelineServiceNS, followerID),
		time.Now().String(),
	)

	// 7. Fan-in recent posts from followee to follower's timeline
	for i := 0; i < 3; i++ {
		postID := wl.NextKeyName()
		timelineEntry := TimelineEntry{
			PostID:    postID,
			UserID:    followeeID,
			Timestamp: time.Now(),
		}
		timelineData, _ := json.Marshal(timelineEntry)
		db.Update(
			ctx,
			"Redis",
			fmt.Sprintf("%v:%v:%v", wl.HomeTimelineServiceNS, followerID, postID),
			string(timelineData),
		)
	}

	db.Commit()
}

// Task 5: UnfollowUser - DeathStarBench: Unfollow a user
func (wl *SocialNetworkWorkload) UnfollowUser(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()

	followerID := wl.NextKeyName()
	followeeID := wl.NextKeyName()

	// 1. Remove from follower's following list (KVRocks)
	_, _ = db.Read(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:%v:following:%v", wl.SocialGraphServiceNS, followerID, followeeID),
	)

	// 2. Remove from followee's followers list (KVRocks)
	_, _ = db.Read(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:%v:followers:%v", wl.SocialGraphServiceNS, followeeID, followerID),
	)

	// 3. Decrement follower count (KVRocks)
	db.Update(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:%v:follower_count", wl.SocialGraphServiceNS, followeeID),
		"-1",
	)

	// 4. Decrement following count (KVRocks)
	db.Update(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:%v:following_count", wl.SocialGraphServiceNS, followerID),
		"-1",
	)

	// 5. Invalidate home timeline cache (Redis)
	db.Update(
		ctx,
		"Redis",
		fmt.Sprintf("%v:%v:cache:invalidate", wl.HomeTimelineServiceNS, followerID),
		time.Now().String(),
	)

	db.Commit()
}

// Task 6: LikePost - DeathStarBench: Like a post
func (wl *SocialNetworkWorkload) LikePost(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()

	userID := wl.NextKeyName()
	postID := wl.NextKeyName()

	// 1. Verify user exists (User Service)
	_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.UserServiceNS, userID))

	// 2. Verify post exists (Post Service)
	_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.PostServiceNS, postID))

	// 3. Add like record (KVRocks for fast writes)
	likeData := map[string]interface{}{
		"user_id":    userID,
		"post_id":    postID,
		"created_at": time.Now(),
	}
	likeDataJSON, _ := json.Marshal(likeData)
	db.Update(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:%v:likes:%v", wl.PostServiceNS, postID, userID),
		string(likeDataJSON),
	)

	// 4. Increment like count in Redis (for fast reads)
	db.Update(
		ctx,
		"Redis",
		fmt.Sprintf("%v:%v:likes", wl.PostServiceNS, postID),
		"1",
	)

	// 5. Update post like count in MongoDB (eventual consistency)
	db.Update(
		ctx,
		"MongoDB2",
		fmt.Sprintf("%v:%v:like_count", wl.PostServiceNS, postID),
		"1",
	)

	// 6. Add to user's liked posts list (KVRocks)
	db.Update(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:%v:liked_posts:%v", wl.UserServiceNS, userID, postID),
		string(likeDataJSON),
	)

	db.Commit()
}

// Task 7: CommentOnPost - DeathStarBench: Comment on a post
func (wl *SocialNetworkWorkload) CommentOnPost(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()

	userID := wl.NextKeyName()
	postID := wl.NextKeyName()
	commentID := wl.NextKeyName()

	// 1. Verify user exists (User Service)
	_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.UserServiceNS, userID))

	// 2. Verify post exists (Post Service)
	_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.PostServiceNS, postID))

	// 3. Create comment (MongoDB)
	comment := wl.generateComment(commentID, postID, userID)
	commentData, _ := json.Marshal(comment)
	db.Update(
		ctx,
		"MongoDB2",
		fmt.Sprintf("%v:comments:%v", wl.PostServiceNS, commentID),
		string(commentData),
	)

	// 4. Add comment to post's comment list (Cassandra - time-ordered)
	commentEntry := map[string]interface{}{
		"comment_id": commentID,
		"user_id":    userID,
		"created_at": time.Now(),
	}
	commentEntryData, _ := json.Marshal(commentEntry)
	db.Update(
		ctx,
		"Cassandra",
		fmt.Sprintf("%v:%v:comments:%v", wl.PostServiceNS, postID, commentID),
		string(commentEntryData),
	)

	// 5. Increment comment count (Redis)
	db.Update(
		ctx,
		"Redis",
		fmt.Sprintf("%v:%v:comments", wl.PostServiceNS, postID),
		"1",
	)

	// 6. Update post comment count (MongoDB)
	db.Update(
		ctx,
		"MongoDB2",
		fmt.Sprintf("%v:%v:comment_count", wl.PostServiceNS, postID),
		"1",
	)

	// 7. Notify post author (stored in KVRocks for fast notification retrieval)
	postAuthorID := wl.NextKeyName()
	notification := map[string]interface{}{
		"type":       "comment",
		"user_id":    userID,
		"post_id":    postID,
		"comment_id": commentID,
		"created_at": time.Now(),
	}
	notificationData, _ := json.Marshal(notification)
	db.Update(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:notifications:%v:%v", wl.UserServiceNS, postAuthorID, commentID),
		string(notificationData),
	)

	db.Commit()
}

// Task 8: GetUserProfile - DeathStarBench: Get user profile with stats
func (wl *SocialNetworkWorkload) GetUserProfile(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()

	userID := wl.NextKeyName()
	viewerID := wl.NextKeyName()

	// 1. Check if profile is cached (Redis)
	_, _ = db.Read(
		ctx,
		"Redis",
		fmt.Sprintf("%v:%v:profile_cache", wl.UserServiceNS, userID),
	)

	// 2. Get user basic info (MongoDB)
	_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.UserServiceNS, userID))

	// 3. Get follower count (KVRocks)
	_, _ = db.Read(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:%v:follower_count", wl.SocialGraphServiceNS, userID),
	)

	// 4. Get following count (KVRocks)
	_, _ = db.Read(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:%v:following_count", wl.SocialGraphServiceNS, userID),
	)

	// 5. Get post count (KVRocks)
	_, _ = db.Read(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:%v:post_count", wl.SocialGraphServiceNS, userID),
	)

	// 6. Get recent posts (Cassandra)
	for i := 0; i < 3; i++ {
		postID := wl.NextKeyName()
		_, _ = db.Read(
			ctx,
			"Cassandra",
			fmt.Sprintf("%v:%v:%v", wl.UserTimelineServiceNS, userID, postID),
		)
	}

	// 7. Check if viewer follows this user (KVRocks)
	_, _ = db.Read(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:%v:following:%v", wl.SocialGraphServiceNS, viewerID, userID),
	)

	// 8. Cache the assembled profile (Redis)
	profileCache := map[string]interface{}{
		"user_id":    userID,
		"cached_at":  time.Now(),
		"expires_at": time.Now().Add(5 * time.Minute),
	}
	cacheData, _ := json.Marshal(profileCache)
	db.Update(
		ctx,
		"Redis",
		fmt.Sprintf("%v:%v:profile_cache", wl.UserServiceNS, userID),
		string(cacheData),
	)

	db.Commit()
}

// Task 9: SearchUsers - DeathStarBench: Search for users
func (wl *SocialNetworkWorkload) SearchUsers(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()

	searcherID := wl.NextKeyName()
	searchQuery := wl.usernames[wl.r.Intn(len(wl.usernames))]

	// 1. Check search cache (Redis)
	_, _ = db.Read(
		ctx,
		"Redis",
		fmt.Sprintf("%v:search:%v", wl.UserServiceNS, searchQuery),
	)

	// 2. Search user index (MongoDB text search simulation)
	for i := 0; i < 3; i++ {
		userID := wl.NextKeyName()
		_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.UserServiceNS, userID))

		// Get user stats (KVRocks)
		_, _ = db.Read(
			ctx,
			"KVRocks",
			fmt.Sprintf("%v:%v:follower_count", wl.SocialGraphServiceNS, userID),
		)
	}

	// 3. Cache search results (Redis)
	searchResults := map[string]interface{}{
		"query":        searchQuery,
		"result_count": 3,
		"cached_at":    time.Now(),
	}
	resultsData, _ := json.Marshal(searchResults)
	db.Update(
		ctx,
		"Redis",
		fmt.Sprintf("%v:search:%v", wl.UserServiceNS, searchQuery),
		string(resultsData),
	)

	// 4. Log search activity (KVRocks)
	searchLog := map[string]interface{}{
		"user_id":   searcherID,
		"query":     searchQuery,
		"timestamp": time.Now(),
	}
	logData, _ := json.Marshal(searchLog)
	db.Update(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:search_log:%v:%v", wl.UserServiceNS, searcherID, time.Now().Unix()),
		string(logData),
	)

	db.Commit()
}

// Task 10: GetRecommendedUsers - DeathStarBench: Get user recommendations
func (wl *SocialNetworkWorkload) GetRecommendedUsers(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()

	userID := wl.NextKeyName()

	// 1. Check recommendation cache (Redis)
	_, _ = db.Read(
		ctx,
		"Redis",
		fmt.Sprintf("%v:recommendations:%v", wl.SocialGraphServiceNS, userID),
	)

	// 2. Get user's following list (KVRocks)
	for i := 0; i < 3; i++ {
		followingID := wl.NextKeyName()
		_, _ = db.Read(
			ctx,
			"KVRocks",
			fmt.Sprintf("%v:%v:following:%v", wl.SocialGraphServiceNS, userID, followingID),
		)

		// Get their followers (friend-of-friend recommendation)
		for j := 0; j < 3; j++ {
			potentialFollowID := wl.NextKeyName()
			_, _ = db.Read(
				ctx,
				"KVRocks",
				fmt.Sprintf(
					"%v:%v:followers:%v",
					wl.SocialGraphServiceNS,
					followingID,
					potentialFollowID,
				),
			)
		}
	}

	// 3. Get trending users (Redis sorted set simulation)
	for i := 0; i < 3; i++ {
		trendingUserID := wl.NextKeyName()
		_, _ = db.Read(
			ctx,
			"Redis",
			fmt.Sprintf("%v:trending:%v", wl.UserServiceNS, trendingUserID),
		)

		// Get user profile
		_, _ = db.Read(ctx, "MongoDB2", fmt.Sprintf("%v:%v", wl.UserServiceNS, trendingUserID))
	}

	// 4. Build and cache recommendations (Redis)
	recommendations := map[string]interface{}{
		"user_id":     userID,
		"recommended": []string{},
		"computed_at": time.Now(),
		"expires_at":  time.Now().Add(1 * time.Hour),
	}
	recData, _ := json.Marshal(recommendations)
	db.Update(
		ctx,
		"Redis",
		fmt.Sprintf("%v:recommendations:%v", wl.SocialGraphServiceNS, userID),
		string(recData),
	)

	db.Commit()
}

func (wl *SocialNetworkWorkload) Load(ctx context.Context, opCount int, db ycsb.DB) {
	txnDB, ok := db.(ycsb.TransactionDB)
	if !ok {
		fmt.Println("The DB does not support transactions")
		return
	}
	if opCount%benconfig.MaxLoadBatchSize != 0 {
		log.Fatalf(
			"opCount should be a multiple of MaxLoadBatchSize, opCount: %d, MaxLoadBatchSize: %d",
			opCount,
			benconfig.MaxLoadBatchSize,
		)
	}

	round := opCount / benconfig.MaxLoadBatchSize
	var aErr error
	for i := 0; i < round; i++ {
		txnDB.Start()
		for j := 0; j < benconfig.MaxLoadBatchSize; j++ {
			key := wl.NextKeyNameFromSequence()

			// Load user data (User Service) - MongoDB
			user := wl.generateUser(key)
			userData, _ := json.Marshal(user)
			txnDB.Insert(
				ctx,
				"MongoDB2",
				fmt.Sprintf("%v:%v", wl.UserServiceNS, key),
				string(userData),
			)

			// Initialize social graph counters (KVRocks)
			txnDB.Insert(
				ctx,
				"KVRocks",
				fmt.Sprintf("%v:%v:follower_count", wl.SocialGraphServiceNS, key),
				"0",
			)
			txnDB.Insert(
				ctx,
				"KVRocks",
				fmt.Sprintf("%v:%v:following_count", wl.SocialGraphServiceNS, key),
				"0",
			)
			txnDB.Insert(
				ctx,
				"KVRocks",
				fmt.Sprintf("%v:%v:post_count", wl.SocialGraphServiceNS, key),
				"0",
			)

			// Load initial posts (Post Service) - MongoDB
			for k := 0; k < 3; k++ {
				postID := fmt.Sprintf("%s_post_%d", key, k)
				post := wl.generatePost(postID, key)
				postData, _ := json.Marshal(post)
				txnDB.Insert(
					ctx,
					"MongoDB2",
					fmt.Sprintf("%v:%v", wl.PostServiceNS, postID),
					string(postData),
				)

				// Add to user timeline (Cassandra)
				timelineEntry := TimelineEntry{
					PostID:    postID,
					UserID:    key,
					Timestamp: time.Now().Add(-time.Duration(k) * time.Hour),
				}
				timelineData, _ := json.Marshal(timelineEntry)
				txnDB.Insert(
					ctx,
					"Cassandra",
					fmt.Sprintf("%v:%v:%v", wl.UserTimelineServiceNS, key, postID),
					string(timelineData),
				)

				// Cache post like count (Redis)
				txnDB.Insert(
					ctx,
					"Redis",
					fmt.Sprintf("%v:%v:likes", wl.PostServiceNS, postID),
					"0",
				)
			}
		}
		err := txnDB.Commit()
		if err != nil {
			aErr = err
			fmt.Printf("Error when committing transaction: %v\n", err)
		}
	}
	if aErr != nil {
		fmt.Printf("Error in Social Network Load: %v\n", aErr)
	}
}

func (wl *SocialNetworkWorkload) Run(ctx context.Context, opCount int, db ycsb.DB) {
	txnDB, ok := db.(ycsb.TransactionDB)
	if !ok {
		fmt.Println("The DB does not support transactions")
		return
	}
	for i := 0; i < opCount; i++ {
		switch wl.NextTask() {
		case 1:
			wl.ComposePost(ctx, txnDB)
		case 2:
			wl.ReadHomeTimeline(ctx, txnDB)
		case 3:
			wl.ReadUserTimeline(ctx, txnDB)
		case 4:
			wl.FollowUser(ctx, txnDB)
		case 5:
			wl.UnfollowUser(ctx, txnDB)
		case 6:
			wl.LikePost(ctx, txnDB)
		case 7:
			wl.CommentOnPost(ctx, txnDB)
		case 8:
			wl.GetUserProfile(ctx, txnDB)
		case 9:
			wl.SearchUsers(ctx, txnDB)
		case 10:
			wl.GetRecommendedUsers(ctx, txnDB)
		default:
			panic("Invalid task")
		}
		thinkingTime := rand.Intn(5) + 5
		time.Sleep(time.Duration(thinkingTime) * time.Millisecond)
	}
}

func (wl *SocialNetworkWorkload) Cleanup() {}

func (wl *SocialNetworkWorkload) NeedPostCheck() bool {
	return true
}

func (wl *SocialNetworkWorkload) NeedRawDB() bool {
	return false
}

func (wl *SocialNetworkWorkload) PostCheck(context.Context, ycsb.DB, chan int) {}

func (wl *SocialNetworkWorkload) DisplayCheckResult() {
	fmt.Printf("Task 1 (Compose Post) count: %v\n", wl.task1Count)
	fmt.Printf("Task 2 (Read Home Timeline) count: %v\n", wl.task2Count)
	fmt.Printf("Task 3 (Read User Timeline) count: %v\n", wl.task3Count)
	fmt.Printf("Task 4 (Follow User) count: %v\n", wl.task4Count)
	fmt.Printf("Task 5 (Unfollow User) count: %v\n", wl.task5Count)
	fmt.Printf("Task 6 (Like Post) count: %v\n", wl.task6Count)
	fmt.Printf("Task 7 (Comment On Post) count: %v\n", wl.task7Count)
	fmt.Printf("Task 8 (Get User Profile) count: %v\n", wl.task8Count)
	fmt.Printf("Task 9 (Search Users) count: %v\n", wl.task9Count)
	fmt.Printf("Task 10 (Get Recommended Users) count: %v\n", wl.task10Count)
}

func (wl *SocialNetworkWorkload) NextTask() int64 {
	wl.mu.Lock()
	defer wl.mu.Unlock()
	idx := wl.taskChooser.Next(wl.r)
	switch idx {
	case 1:
		wl.task1Count++
	case 2:
		wl.task2Count++
	case 3:
		wl.task3Count++
	case 4:
		wl.task4Count++
	case 5:
		wl.task5Count++
	case 6:
		wl.task6Count++
	case 7:
		wl.task7Count++
	case 8:
		wl.task8Count++
	case 9:
		wl.task9Count++
	case 10:
		wl.task10Count++
	default:
		panic("Invalid task")
	}
	return idx
}

func (wl *SocialNetworkWorkload) RandomValue() string {
	wl.mu.Lock()
	defer wl.mu.Unlock()
	value := wl.r.Intn(100000)
	return util.ToString(value)
}
