package main


type (
	cacheDb struct {
		db map[string]string
	}
)


func newCacheDb() *cacheDb {
	return &cacheDb{
		db: map[string]string{},
	}
}

func (c *cacheDb) Get(key string ) string {
	return key
}