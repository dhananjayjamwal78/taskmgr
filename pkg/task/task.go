package task
import "time"
type Task struct {
    ID      uint64    `json:"id"`     
    Text    string    `json:"text"`   
	Created time.Time `json:"created"` 
    Due     time.Time `json:"due"`     
    Done    bool      `json:"done"`  
}