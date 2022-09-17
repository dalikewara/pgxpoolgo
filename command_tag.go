package pgxpoolgo

type CommandTag interface {
	RowsAffected() int64
	String() string
	Insert() bool
	Update() bool
	Delete() bool
	Select() bool
}
