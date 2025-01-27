package connection

import (
	"fmt"

	"github.com/epos-eu/converter-service/dao/model"
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

func GetPluginRelationByOperationId(operationId string) ([]model.PluginRelation, error) {
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
