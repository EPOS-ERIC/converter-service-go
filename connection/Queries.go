package connection

import (
	"fmt"

	"github.com/epos-eu/converter-service/orms"
)

func GetPlugins() ([]orms.Plugin, error) {
	db, err := Connect()
	if err != nil {
		return nil, err
	}
	// Select all users.
	var listOfPlugins []orms.Plugin
	err = db.Model(&listOfPlugins).Find(&listOfPlugins).Error
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
	// Select all users.
	var listOfPluginRelations []orms.PluginRelations
	err = db.Model(&listOfPluginRelations).Find(&listOfPluginRelations).Error
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
	err = db.Model(&plugin).Where("id = ?", id).First(&plugin).Error
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

	// Get the operation by id
	var operation orms.Operation
	err = db.Model(&operation).Where("uid = ?", operationId).First(&operation).Error
	if err != nil {
		return nil, err
	}

	// Get the plugin relations by operationInstanceId
	var listOfPluginRelations []orms.PluginRelations
	err = db.Model(&listOfPluginRelations).Where("relation_id = ?", operation.InstanceID).Find(&listOfPluginRelations).Error
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
	err = db.Model(&plugin).Where("id = ?", pluginId).First(&plugin).Error
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
	err = db.Model(plugin).
		Where("id = ?", id).
		Update("enabled", enable).
		Error
	if err != nil {
		return err
	}
	return nil
}
