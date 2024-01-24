package couchdb

// NewDefaultCouchDBConnection creates a new instance of CouchDBConnection with default settings.
// It establishes a connection to the CouchDBDB server running on localhost:27017.
// Returns a pointer to the CouchDBConnection.
func NewDefaultConnection() *CouchDBConnection {
	conn := NewCouchDBConnection(&ConnectionOptions{
		Address: "http://localhost:5984",
		DBName:  "oreo",
	})
	err := conn.Connect()
	if err != nil {
		panic(err)
	}
	return conn
}
