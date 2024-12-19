package orms

type Operation struct {
	tableName          struct{} `gorm:"operation,alias:operation"`
	UID                string   `gorm:"uid" gorm:"column:uid;primaryKey"`
	Method             string   `gorm:"method" gorm:"column:method"`
	Template           string   `gorm:"template" gorm:"column:template"`
	SupportedOperation string   `gorm:"supportedoperation" gorm:"column:supported_operation"`
	FileProvenance     string   `gorm:"fileprovenance" gorm:"column:file_provenance"`
	InstanceID         string   `gorm:"instance_id" gorm:"column:instance_id"`
	MetaID             string   `gorm:"meta_id" gorm:"column:meta_id"`
	InstanceChangedID  string   `gorm:"instance_changed_id" gorm:"column:instance_changed_id"`
	ChangeTimestamp    string   `gorm:"change_timestamp" gorm:"column:change_timestamp"`
	Operation          string   `gorm:"operation" gorm:"column:operation"`
	EditorMetaID       string   `gorm:"editor_meta_id" gorm:"column:editor_meta_id"`
	ChangeComment      string   `gorm:"change_comment" gorm:"column:change_comment"`
	ReviewerMetaID     string   `gorm:"reviewer_meta_id" gorm:"column:reviewer_meta_id"`
	ReviewComment      string   `gorm:"review_comment" gorm:"column:review_comment"`
	Version            string   `gorm:"version" gorm:"column:version"`
	State              string   `gorm:"state" gorm:"column:state"`
	ToBeDeleted        bool     `gorm:"to_be_deleted" gorm:"column:to_be_deleted"`
}

func (Operation) TableName() string {
	return "operation" // Replace this with your actual table name
}
