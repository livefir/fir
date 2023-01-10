package eventstate

type Type string

const (
	OK      Type = "ok"
	Error   Type = "error"
	Pending Type = "pending"
	Done    Type = "done"
)
