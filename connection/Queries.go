package connection

import (
	"fmt"

	"github.com/epos-eu/converter-service/dao/model"
	"github.com/google/uuid"
)

func GetPlugins() ([]model.Plugin, error) {
	db, err := Connect()
	if err != nil {
		return nil, err
	}
	// Select all users.
	var listOfPlugins []model.Plugin
	err = db.Model(&listOfPlugins).Find(&listOfPlugins).Error
	if err != nil {
		return nil, err
	}
	return listOfPlugins, nil
}

func GetPluginRelation() ([]model.PluginRelation, error) {
	db, err := Connect()
	if err != nil {
		return nil, err
	}
	// Select all users.
	var listOfPluginRelation []model.PluginRelation
	err = db.Model(&listOfPluginRelation).Find(&listOfPluginRelation).Error
	if err != nil {
		panic(err)
	}
	return listOfPluginRelation, nil
}

func GetPluginRelationById(id string) (model.PluginRelation, error) {
	var plugin model.PluginRelation
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

func GetPluginRelationsByOperationId(operationId string) ([]model.PluginRelation, error) {
	db, err := Connect()
	if err != nil {
		return nil, err
	}

	// TODO: ceck if this is necessary
	// // Get the operation by id
	// var operation model.Operation
	// err = db.Model(&operation).Where("uid = ?", operationId).First(&operation).Error
	// if err != nil {
	// 	return nil, err
	// }

	// Get the plugin relations by operationInstanceId
	var listOfPluginRelation []model.PluginRelation
	err = db.Model(&listOfPluginRelation).Where("relation_id = ?", operationId).Find(&listOfPluginRelation).Error
	if err != nil {
		return nil, err
	}
	if len(listOfPluginRelation) == 0 {
		return nil, fmt.Errorf("eror: found 0 plugins related to OperationId: %s", operationId)
	}
	return listOfPluginRelation, nil
}

func GetPluginById(pluginId string) (model.Plugin, error) {
	var plugin model.Plugin
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
	plugin := &model.Plugin{}

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

func UpdatePlugin(id string, plugin model.Plugin) error {
	db, err := Connect()
	if err != nil {
		return err
	}

	// Find the existing plugin record by ID
	var existing model.Plugin
	err = db.First(&existing, "id = ?", id).Error
	if err != nil {
		return err
	}

	// Update the existing plugin record with the new data
	err = db.Model(&existing).Updates(plugin).Error
	if err != nil {
		return err
	}

	return nil
}

func DeletePlugin(id string) (plugin model.Plugin, err error) {
	db, err := Connect()
	if err != nil {
		return plugin, err
	}

	// Retrieve the plugin to be deleted
	err = db.First(&plugin, "id = ?", id).Error
	if err != nil {
		return plugin, err
	}

	// Delete the plugin record
	err = db.Delete(&plugin).Error
	if err != nil {
		return plugin, err
	}

	return plugin, nil
}

func CreatePlugin(plugin model.Plugin) (model.Plugin, error) {
	db, err := Connect()
	if err != nil {
		return plugin, err
	}

	// Generate an id for the plugin
	plugin.ID = uuid.New().String()
	err = db.Create(&plugin).Error
	if err != nil {
		return plugin, err
	}

	return plugin, nil
}

func UpdatePluginRelation(id string, relation model.PluginRelation) error {
	db, err := Connect()
	if err != nil {
		return err
	}

	// Find the existing plugin record by ID
	var existing model.Plugin
	err = db.First(&existing, "id = ?", id).Error
	if err != nil {
		return err
	}

	// Update the existing plugin record with the new data
	err = db.Model(&existing).Updates(relation).Error
	if err != nil {
		return err
	}

	return nil
}

func DeletePluginRelation(id string) (relation model.PluginRelation, err error) {
	db, err := Connect()
	if err != nil {
		return relation, err
	}

	// Retrieve the plugin to be deleted
	err = db.First(&relation, "id = ?", id).Error
	if err != nil {
		return relation, err
	}

	// Delete the plugin record
	err = db.Delete(&relation).Error
	if err != nil {
		return relation, err
	}

	return relation, nil
}

func CreatePluginRelation(relation model.PluginRelation) (model.PluginRelation, error) {
	db, err := Connect()
	if err != nil {
		return relation, err
	}

	// Generate an id for the plugin
	relation.ID = uuid.New().String()
	err = db.Create(&relation).Error
	if err != nil {
		return relation, err
	}

	return relation, nil
}
