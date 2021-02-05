/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package repository

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type GitOpsConfigRepository interface {
	CreateGitOpsConfig(model *GitOpsConfig) (*GitOpsConfig, error)
	UpdateGitOpsConfig(model *GitOpsConfig) error
	GetGitOpsConfigById(id int) (*GitOpsConfig, error)
	GetAllGitOpsConfig() ([]*GitOpsConfig, error)
}

type GitOpsConfigRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

type GitOpsConfig struct {
	tableName     struct{} `sql:"gitops_config" pg:",discard_unknown_columns"`
	Id            int      `sql:"id,pk"`
	Provider      string   `sql:"provider"`
	Username      string   `sql:"username"`
	Token         string   `sql:"token"`
	GitLabGroupId string   `sql:"gitlab_group_id"`
	GitHubOrgId   string   `sql:"github_org_id"`
	Host          string   `sql:"host"`
	Active        bool     `sql:"active,notnull"`
	models.AuditLog
}

func NewGitOpsConfigRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *GitOpsConfigRepositoryImpl {
	return &GitOpsConfigRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl *GitOpsConfigRepositoryImpl) CreateGitOpsConfig(model *GitOpsConfig) (*GitOpsConfig, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.logger.Error(err)
		return model, err
	}
	return model, nil
}
func (impl *GitOpsConfigRepositoryImpl) UpdateGitOpsConfig(model *GitOpsConfig) error {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}
func (impl *GitOpsConfigRepositoryImpl) GetGitOpsConfigById(id int) (*GitOpsConfig, error) {
	var model GitOpsConfig
	err := impl.dbConnection.Model(&model).Where("id = ?", id).Where("active = ?", true).Select()
	return &model, err
}
func (impl *GitOpsConfigRepositoryImpl) GetAllGitOpsConfig() ([]*GitOpsConfig, error) {
	var userModel []*GitOpsConfig
	err := impl.dbConnection.Model(&userModel).Where("active = ?", true).Order("updated_on desc").Select()
	return userModel, err
}
