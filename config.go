package blueregister

type Config struct {
	Prefix string `bson:"_id" json:"_id"`
	Pid    int    `json:"pid" bson:"pid"`
}
