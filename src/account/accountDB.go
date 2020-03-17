package account

import (
	"errors"

	"github.com/gomodule/redigo/redis"
)

type memorydb struct {
	Expires map[string]bool
}

type redisdb struct {
	cache redis.Conn
}

func (rdb redisdb) connect() {
	log.Debug("rdb: connecting localhost database...")
	conn, err := redis.DialURL("redis://localhost")
	if err != nil {
		rdb.cache = nil
	}
	rdb.cache = conn
}

func (rdb redisdb) disconnect() {

}

func (rdb redisdb) save(uuid string, expires bool) error {
	log.Debugf("rdb: setex \"%s\" to database", uuid)
	_, err := rdb.cache.Do("SETEX", uuid, "300", expires)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (rdb redisdb) get(uuid string) (bool, error) {
	log.Debugf("rdb: get \"%s\" from database", uuid)
	_, err := rdb.cache.Do("GET", uuid)
	if err != nil {
		log.Error(err.Error())
		return false, err
	}
	return true, nil
}

func (rdb redisdb) del(uuid string) error {
	log.Debugf("rdb: del \"%s\" from database", uuid)
	_, err := rdb.cache.Do("DEL", uuid)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (db memorydb) connect() {

}
func (db memorydb) disconnect() {

}

func (db memorydb) save(uuid string, expires bool) error {
	log.Debugf("mdb: set \"%s\" to database", uuid)
	if uuid == "" {
		log.Error("uuid is empty, exit....")
		return errors.New("uuid is empty")
	}
	db.Expires[uuid] = expires
	return nil
}

func (db memorydb) get(uuid string) (bool, error) {
	log.Debugf("mdb: get \"%s\" from database", uuid)
	return db.Expires[uuid], nil
}

func (db memorydb) del(uuid string) error {
	log.Debugf("mdb: del \"%s\" from database", uuid)
	delete(db.Expires, uuid)
	return nil
}

type DB interface {
	connect()
	disconnect()
	save(string, bool) error
	get(string) (bool, error)
	del(string) error
}

// func (a *account) Login() error {
// 	if a.Name != string(user) || a.Passwd != string(passwd) {
// 		log.Error("Login failed")
// 		return errors.New("Login failed")
// 	}
// 	var err error
// 	a.uuid, err = uuid.NewUUID()
// 	if err != nil {
// 		log.Error(err.Error())
// 		return err
// 	}
// 	a.Db.save(a.uuid.String(), true)
// }

// func (a *account) Logout() {
// 	a.Db.del(a.uuid.String())
// }

// func (a *account) Verify() bool {
// 	res, err := a.Db.get(a.uuid.String())
// 	if err != nil {
// 		log.Error(err.Error())
// 		return false
// 	}
// 	return res
// }
