package db

import (
	"fmt"

	"github.com/epos-eu/converter-service/dao/model"
	"gorm.io/gorm/clause"
)

func GetPlugins() ([]model.Plugin, error) {
	db := Get()

	var listOfPlugins []model.Plugin
	err := db.Model(&listOfPlugins).Find(&listOfPlugins).Error
	if err != nil {
		return nil, err
	}
	return listOfPlugins, nil
}

func GetAllPluginRelations() ([]model.PluginRelation, error) {
	db := Get()

	var listOfPluginRelation []model.PluginRelation
	err := db.Model(&listOfPluginRelation).Find(&listOfPluginRelation).Error
	if err != nil {
		return nil, err
	}
	return listOfPluginRelation, nil
}

func GetPluginRelationForEnabledPlugins() ([]model.PluginRelation, error) {
	db := Get()

	var listOfPluginRelation []model.PluginRelation
	// Join the plugins table and filter where plugins.enabled and plugins.installed are true.
	err := db.
		Joins("JOIN converter_catalogue.plugin ON converter_catalogue.plugin.id = converter_catalogue.plugin_relations.plugin_id").
		Where("converter_catalogue.plugin.enabled = ? AND converter_catalogue.plugin.installed = ?", true, true).
		Find(&listOfPluginRelation).Error
	if err != nil {
		return nil, err
	}
	return listOfPluginRelation, nil
}

func GetPluginRelationByID(id string) (model.PluginRelation, error) {
	var plugin model.PluginRelation
	db := Get()

	err := db.Model(&plugin).Where("id = ?", id).First(&plugin).Error
	if err != nil {
		return plugin, err
	}
	return plugin, nil
}

func GetPluginRelationsByRelationID(relationID string) ([]model.PluginRelation, error) {
	db := Get()

	var relations []model.PluginRelation
	err := db.
		Joins("JOIN converter_catalogue.plugin ON converter_catalogue.plugin.id = converter_catalogue.plugin_relations.plugin_id").
		Where("converter_catalogue.plugin_relations.relation_id = ?", relationID).
		Find(&relations).Error
	if err != nil {
		return nil, err
	}
	return relations, nil
}

func GetPluginByID(pluginID string) (model.Plugin, error) {
	var plugin model.Plugin
	db := Get()

	err := db.Model(&plugin).Where("id = ?", pluginID).First(&plugin).Error
	if err != nil {
		return plugin, err
	}
	return plugin, nil
}

func EnablePlugin(id string, enable bool) error {
	plugin := &model.Plugin{}
	db := Get()

	err := db.Model(plugin).
		Where("id = ?", id).
		Update("enabled", enable).
		Error
	if err != nil {
		return err
	}
	return nil
}

// UpdatePlugin needs the id of the plugin to be set
func UpdatePlugin(plugin model.Plugin) error {
	if plugin.ID == "" {
		return fmt.Errorf("plugin id not set, can't update a plugin without an ID: %+v", plugin)
	}
	db := Get()

	// Update the existing plugin record with the new data
	err := db.Model(&plugin).Select("*").Updates(plugin).Error
	if err != nil {
		return err
	}

	return nil
}

func DeletePlugin(id string) (plugin model.Plugin, err error) {
	db := Get()

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

// CreatePlugin creates a new plugin in the db.
// If the plugin already exists nothing is done  and the original one is returned
func CreatePlugin(plugin model.Plugin) (model.Plugin, error) {
	db := Get()

	res := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&plugin)
	if res.Error != nil {
		return plugin, res.Error
	}

	if res.RowsAffected == 0 {
		// reset the id so that gorm doesn't include it in the where
		plugin.ID = ""
		var existing model.Plugin
		err := db.
			Where(&plugin).
			First(&existing).
			Error
		if err != nil {
			return plugin, err
		}
		return existing, nil
	}

	// newly inserted
	return plugin, nil
}

func UpdatePluginRelation(relation model.PluginRelation) error {
	if relation.ID == "" {
		return fmt.Errorf("the id of the relation is not set, can't update a relation without an ID: %+v", relation)
	}
	db := Get()

	// Update the existing relation record with the new data
	err := db.Model(&relation).Select("*").Updates(relation).Error
	if err != nil {
		return err
	}

	return nil
}

func DeletePluginRelation(id string) (relation model.PluginRelation, err error) {
	db := Get()

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

// CreatePluginRelation creates a new plugin relation in the db.
// If the relation already exists nothing is done and the original one is returned
func CreatePluginRelation(relation model.PluginRelation) (model.PluginRelation, error) {
	db := Get()

	res := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&relation)
	if res.Error != nil {
		return relation, res.Error
	}

	if res.RowsAffected == 0 {
		// reset the id so that gorm doesn't include it in the where
		relation.ID = ""
		var existing model.PluginRelation
		err := db.
			Where(&relation).
			First(&existing).
			Error
		if err != nil {
			return relation, err
		}
		return existing, nil
	}

	return relation, nil
}
