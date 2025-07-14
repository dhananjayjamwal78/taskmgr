package store

import (
    "encoding/binary"
    "encoding/json"
    "errors"
    "time"

    bolt "go.etcd.io/bbolt"
    "github.com/dhananjayjamwal78/taskmgr/pkg/task"
)

// Store wraps the BoltDB instance and provides CRUD for Task.
type Store struct {
    db *bolt.DB
}

const bucketName = "tasks"

// NewBoltStore opens (or creates) the DB at dbPath and ensures the bucket exists.
func NewBoltStore(dbPath string) (*Store, error) {
    db, err := bolt.Open(dbPath, 0600, nil)
    if err != nil {
        return nil, err
    }
    // ensure bucket
    if err := db.Update(func(tx *bolt.Tx) error {
        _, err := tx.CreateBucketIfNotExists([]byte(bucketName))
        return err
    }); err != nil {
        db.Close()
        return nil, err
    }
    return &Store{db: db}, nil
}

// AddTask inserts a new Task with auto-incremented ID.
func (s *Store) AddTask(text string, due time.Time) (uint64, error) {
    var id uint64
    err := s.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(bucketName))
        seq, err := b.NextSequence()
        if err != nil {
            return err
        }
        id = seq

        t := task.Task{
            ID:      id,
            Text:    text,
            Created: time.Now(),
            Due:     due,
            Done:    false,
        }
        buf, err := json.Marshal(t)
        if err != nil {
            return err
        }
        return b.Put(itob(id), buf)
    })
    return id, err
}

// ListTasks returns all tasks; if includeDone is false, only pending tasks.
func (s *Store) ListTasks(includeDone bool) ([]task.Task, error) {
    var tasks []task.Task
    err := s.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(bucketName))
        return b.ForEach(func(k, v []byte) error {
            var t task.Task
            if err := json.Unmarshal(v, &t); err != nil {
                return err
            }
            if !includeDone && t.Done {
                return nil
            }
            tasks = append(tasks, t)
            return nil
        })
    })
    return tasks, err
}

// MarkDone sets Done=true on the Task with given ID.
func (s *Store) MarkDone(id uint64) error {
    return s.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(bucketName))
        data := b.Get(itob(id))
        if data == nil {
            return errors.New("task not found")
        }
        var t task.Task
        if err := json.Unmarshal(data, &t); err != nil {
            return err
        }
        t.Done = true
        buf, err := json.Marshal(t)
        if err != nil {
            return err
        }
        return b.Put(itob(id), buf)
    })
}

// DeleteTask removes the Task with given ID.
func (s *Store) DeleteTask(id uint64) error {
    return s.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(bucketName))
        return b.Delete(itob(id))
    })
}

// helper: uint64 ↔ big-endian []byte
func itob(v uint64) []byte {
    buf := make([]byte, 8)
    binary.BigEndian.PutUint64(buf, v)
    return buf
}
// Close cleanly closes the underlying BoltDB.
func (s *Store) Close() error {
    return s.db.Close()
}

