package orms

type Operation struct {
	tableName           struct{} `pg:"operation,alias:operation"`
	Uid                 string
	Method              string
	Template            string
	Supportedoperation  string
	Fileprovenance      string
	Instance_id         string
	Meta_id             string
	Instance_changed_id string
	Change_timestamp    string
	Operation           string
	Editor_meta_id      string
	Change_comment      string
	Reviewer_meta_id    string
	Review_comment      string
	Version             string
	State               string
	To_be_deleted       bool
}
