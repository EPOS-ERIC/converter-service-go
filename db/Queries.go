package db

import (
	"fmt"

	"github.com/epos-eu/converter-service/dao/model"
)

func GetPlugins() ([]model.Plugin, error) {
	db, err := Connect()
	if err != nil {
		return nil, err
	}
	var listOfPlugins []model.Plugin
	err = db.Model(&listOfPlugins).Find(&listOfPlugins).Error
	if err != nil {
		return nil, err
	}
	return listOfPlugins, nil
}

func GetAllPluginRelations() ([]model.PluginRelation, error) {
	db, err := Connect()
	if err != nil {
		return nil, err
	}
	var listOfPluginRelation []model.PluginRelation
	err = db.Model(&listOfPluginRelation).Find(&listOfPluginRelation).Error
	if err != nil {
		return nil, err
	}
	return listOfPluginRelation, nil
}

func GetPluginRelationForEnabledPlugins() ([]model.PluginRelation, error) {
	db, err := Connect()
	if err != nil {
		return nil, err
	}
	var listOfPluginRelation []model.PluginRelation
	// Join the plugins table and filter where plugins.enabled and plugins.installed are true.
	err = db.
		Joins("JOIN plugin ON plugin.id = plugin_relations.plugin_id").
		Where("plugin.enabled = ? AND plugin.installed = ?", true, true).
		Find(&listOfPluginRelation).Error
	if err != nil {
		return nil, err
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

// the id of the plugin needs to be set
func UpdatePlugin(plugin model.Plugin) error {
	if plugin.ID == "" {
		return fmt.Errorf("plugin id not set, can't update a plugin without an ID: %+v", plugin)
	}
	db, err := Connect()
	if err != nil {
		return err
	}

	// Update the existing plugin record with the new data
	err = db.Model(&plugin).Select("*").Updates(plugin).Error
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

	err = db.Create(&plugin).Error
	if err != nil {
		return plugin, err
	}

	return plugin, nil
}

func UpdatePluginRelation(relation model.PluginRelation) error {
	if relation.ID == "" {
		return fmt.Errorf("the id of the relation is not set, can't update a relation without an ID: %+v", relation)
	}
	db, err := Connect()
	if err != nil {
		return err
	}

	// Update the existing relation record with the new data
	err = db.Model(&relation).Select("*").Updates(relation).Error
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

	err = db.First(&relation, "id = ?", id).Error
	if err != nil {
		return relation, err
	}

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

	err = db.Create(&relation).Error
	if err != nil {
		return relation, err
	}

	return relation, nil
}
