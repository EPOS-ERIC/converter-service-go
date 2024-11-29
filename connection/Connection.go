package connection

import (
	"fmt"
	"os"

	"github.com/epos-eu/converter-service/orms"
	"github.com/go-pg/pg/v10"
)

func Connect() (*pg.DB, error) {
	conn := ""

	if val, res := os.LookupEnv("POSTGRESQL_CONNECTION_STRING"); res == true {
		conn = val
	} else {
		return nil, fmt.Errorf("POSTGRESQL_CONNECTION_STRING is not set")
	}

	opt, err := pg.ParseURL(conn)
	if err != nil {
		return nil, err
	}
	db := pg.Connect(opt)

	return db, nil
}

//func GetSoftwareSourceCodes() []orms.SoftwareSourceCode {
//	db := Connect()
//	defer db.Close()
//	// Select all users.
//	var listOfSoftwareSourceCodes []orms.SoftwareSourceCode
//	err := db.Model(&listOfSoftwareSourceCodes).Where("state = ?", "PUBLISHED").Where("uid ILIKE '%' || ? || '%'", "plugin").Select()
//	if err != nil {
//		panic(err)
//	}
//	return listOfSoftwareSourceCodes
//}

//func GetSoftwareApplications() []orms.SoftwareApplication {
//	db := Connect()
//	defer db.Close()
//	// Select all users.
//	var listOfSoftwareApplications []orms.SoftwareApplication
//	err := db.Model(&listOfSoftwareApplications).Where("state = ?", "PUBLISHED").Where("uid ILIKE '%' || ? || '%'", "plugin").Select()
//	if err != nil {
//		panic(err)
//	}
//	return listOfSoftwareApplications
//}

//func GetSoftwareApplicationsOperations() []orms.SoftwareApplicationOperation {
//	db := Connect()
//	defer db.Close()
//	// Select all users.
//	var listOfSoftwareApplicationsOperations []orms.SoftwareApplicationOperation
//	err := db.Model(&listOfSoftwareApplicationsOperations).Select()
//	if err != nil {
//		panic(err)
//	}
//	return listOfSoftwareApplicationsOperations
//}

func GetPlugins() ([]orms.Plugin, error) {
	db, err := Connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	// Select all users.
	var listOfPlugins []orms.Plugin
	err = db.Model(&listOfPlugins).Select()
	if err != nil {
		return nil, err
	}
	return listOfPlugins, nil
}

func GetPluginRelations() ([]orms.PluginRelations, error) {
	db, err := Connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	// Select all users.
	var listOfPluginRelations []orms.PluginRelations
	err = db.Model(&listOfPluginRelations).Select()
	if err != nil {
		panic(err)
	}
	return listOfPluginRelations, nil
}

func GetPluginRelationsById(id string) (orms.PluginRelations, error) {
	var plugin orms.PluginRelations
	db, err := Connect()
	if err != nil {
		return plugin, err
	}
	defer db.Close()
	err = db.Model(&plugin).Where("id = ?", id).Select()
	if err != nil {
		return plugin, err
	}
	return plugin, nil
}

func GetPluginRelationsByOperationId(operationId string) ([]orms.PluginRelations, error) {
	db, err := Connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Get the operation by id
	var operation orms.Operation
	err = db.Model(&operation).Where("uid = ?", operationId).Select()
	if err != nil {
		return nil, err
	}

	// Get the plugin relations by operationInstanceId
	var listOfPluginRelations []orms.PluginRelations
	err = db.Model(&listOfPluginRelations).Where("relation_id = ?", operation.Instance_id).Select()
	if err != nil {
		return nil, err
	}
	if len(listOfPluginRelations) == 0 {
		return nil, fmt.Errorf("eror: found 0 plugins related to OperationId: %s", operationId)
	}
	return listOfPluginRelations, nil
}

func GetPluginById(pluginId string) (orms.Plugin, error) {
	var plugin orms.Plugin
	db, err := Connect()
	if err != nil {
		return plugin, err
	}
	defer db.Close()
	err = db.Model(&plugin).Where("id = ?", pluginId).Select()
	if err != nil {
		return plugin, err
	}
	return plugin, nil
}

func EnablePlugin(id string, enable bool) error {
	plugin := &orms.Plugin{}

	db, err := Connect()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Model(plugin).
		Set("enabled = ?", enable).
		Where("id = ?", id).
		Update()
	if err != nil {
		return err
	}
	return nil
}
