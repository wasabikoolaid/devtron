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

package appStoreRepository

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type InstalledAppRepository interface {
	CreateInstalledApp(model *InstalledApps, tx *pg.Tx) (*InstalledApps, error)
	CreateInstalledAppVersion(model *InstalledAppVersions, tx *pg.Tx) (*InstalledAppVersions, error)
	UpdateInstalledApp(model *InstalledApps, tx *pg.Tx) (*InstalledApps, error)
	UpdateInstalledAppVersion(model *InstalledAppVersions, tx *pg.Tx) (*InstalledAppVersions, error)
	GetInstalledApp(id int) (*InstalledApps, error)
	GetInstalledAppVersion(id int) (*InstalledAppVersions, error)
	GetInstalledAppVersionAny(id int) (*InstalledAppVersions, error)
	GetAllInstalledApps(filter *appStoreBean.AppStoreFilter) ([]InstalledAppsWithChartDetails, error)
	GetAllIntalledAppsByAppStoreId(appStoreId int) ([]InstalledAppAndEnvDetails, error)
	GetAllInstalledAppsByChartRepoId(chartRepoId int) ([]InstalledAppAndEnvDetails, error)
	GetInstalledAppVersionByInstalledAppIdAndEnvId(installedAppId int, envId int) (*InstalledAppVersions, error)
	GetInstalledAppVersionByAppStoreId(appStoreId int) ([]*InstalledAppVersions, error)

	DeleteInstalledApp(model *InstalledApps) (*InstalledApps, error)
	DeleteInstalledAppVersion(model *InstalledAppVersions) (*InstalledAppVersions, error)
	GetInstalledAppVersionByInstalledAppId(id int) ([]*InstalledAppVersions, error)
	GetConnection() (dbConnection *pg.DB)
	GetInstalledAppVersionByInstalledAppIdMeta(appStoreApplicationId int) ([]*InstalledAppVersions, error)
	GetClusterComponentByClusterId(clusterId int) ([]*InstalledApps, error) //unused
	GetClusterComponentByClusterIds(clusterIds []int) ([]*InstalledApps, error) //unused
	GetInstalledAppVersionByAppIdAndEnvId(appId int, envId int) (*InstalledAppVersions, error)
	GetInstalledAppVersionByClusterIds(clusterIds []int) ([]*InstalledAppVersions, error) //unused
	GetInstalledAppVersionByClusterIdsV2(clusterIds []int) ([]*InstalledAppVersions, error)
	GetInstalledApplicationByClusterIdAndNamespaceAndAppName(clusterId int, namespace string, appName string) (*InstalledApps, error)
}

type InstalledAppRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewInstalledAppRepositoryImpl(Logger *zap.SugaredLogger, dbConnection *pg.DB) *InstalledAppRepositoryImpl {
	return &InstalledAppRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

type InstalledApps struct {
	TableName     struct{}                              `sql:"installed_apps" pg:",discard_unknown_columns"`
	Id            int                                   `sql:"id,pk"`
	AppId         int                                   `sql:"app_id,notnull"`
	EnvironmentId int                                   `sql:"environment_id,notnull"`
	Active        bool                                  `sql:"active, notnull"`
	Status        appStoreBean.AppstoreDeploymentStatus `sql:"status"`
	App           app.App
	Environment   repository.Environment
	sql.AuditLog
}

type InstalledAppVersions struct {
	TableName                    struct{} `sql:"installed_app_versions" pg:",discard_unknown_columns"`
	Id                           int      `sql:"id,pk"`
	InstalledAppId               int      `sql:"installed_app_id,notnull"`
	AppStoreApplicationVersionId int      `sql:"app_store_application_version_id,notnull"`
	ValuesYaml                   string   `sql:"values_yaml_raw"`
	Active                       bool     `sql:"active, notnull"`
	ReferenceValueId             int      `sql:"reference_value_id"`
	ReferenceValueKind           string   `sql:"reference_value_kind"`
	sql.AuditLog
	InstalledApp               InstalledApps
	AppStoreApplicationVersion appStoreDiscoverRepository.AppStoreApplicationVersion
}

type InstalledAppsWithChartDetails struct {
	AppStoreApplicationName      string    `json:"app_store_application_name"`
	ChartRepoName                string    `json:"chart_repo_name"`
	AppName                      string    `json:"app_name"`
	EnvironmentName              string    `json:"environment_name"`
	InstalledAppVersionId        int       `json:"installed_app_version_id"`
	AppStoreApplicationVersionId int       `json:"app_store_application_version_id"`
	Icon                         string    `json:"icon"`
	Readme                       string    `json:"readme"`
	CreatedOn                    time.Time `json:"created_on"`
	UpdatedOn                    time.Time `json:"updated_on"`
	Id                           int       `json:"id"`
	EnvironmentId                int       `json:"environment_id"`
	Deprecated                   bool      `json:"deprecated"`
	ClusterName                  string    `json:"clusterName"`
	Namespace                    string    `json:"namespace"`
	TeamId                       int       `json:"teamId"`
	ClusterId                    int       `json:"clusterId"`
}

type InstalledAppAndEnvDetails struct {
	EnvironmentName              string    `json:"environment_name"`
	EnvironmentId                int       `json:"environment_id"`
	AppName                      string    `json:"app_name"`
	AppOfferingMode              string    `json:"appOfferingMode"`
	UpdatedOn                    time.Time `json:"updated_on"`
	EmailId                      string    `json:"email_id"`
	InstalledAppVersionId        int       `json:"installed_app_version_id"`
	InstalledAppId               int       `json:"installed_app_id"`
	AppStoreApplicationVersionId int       `json:"app_store_application_version_id"`
}

func (impl InstalledAppRepositoryImpl) CreateInstalledApp(model *InstalledApps, tx *pg.Tx) (*InstalledApps, error) {
	err := tx.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl InstalledAppRepositoryImpl) CreateInstalledAppVersion(model *InstalledAppVersions, tx *pg.Tx) (*InstalledAppVersions, error) {
	err := tx.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl InstalledAppRepositoryImpl) UpdateInstalledApp(model *InstalledApps, tx *pg.Tx) (*InstalledApps, error) {
	err := tx.Update(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl InstalledAppRepositoryImpl) UpdateInstalledAppVersion(model *InstalledAppVersions, tx *pg.Tx) (*InstalledAppVersions, error) {
	err := tx.Update(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl InstalledAppRepositoryImpl) GetInstalledApp(id int) (*InstalledApps, error) {
	model := &InstalledApps{}
	err := impl.dbConnection.Model(model).
		Column("installed_apps.*", "App", "Environment").
		Where("installed_apps.id = ?", id).Where("installed_apps.active = true").Select()
	return model, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppVersionByAppStoreId(appStoreId int) ([]*InstalledAppVersions, error) {
	var model []*InstalledAppVersions
	err := impl.dbConnection.Model(&model).
		Column("installed_app_versions.*", "InstalledApp", "InstalledApp.App", "InstalledApp.Environment", "AppStoreApplicationVersion", "AppStoreApplicationVersion.AppStore").
		Column("AppStoreApplicationVersion.AppStore.ChartRepo").
		Where("app_store_application_version.app_store_id = ?", appStoreId).
		Where("installed_app_versions.active = true").Select()
	return model, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppVersionByInstalledAppIdMeta(installedAppId int) ([]*InstalledAppVersions, error) {
	var model []*InstalledAppVersions
	err := impl.dbConnection.Model(&model).
		Column("installed_app_versions.*", "InstalledApp", "InstalledApp.App", "InstalledApp.Environment", "AppStoreApplicationVersion", "AppStoreApplicationVersion.AppStore").
		Column("AppStoreApplicationVersion.AppStore.ChartRepo").
		Where("installed_app_versions.installed_app_id = ?", installedAppId).
		Where("installed_app_versions.active = true").Select()
	return model, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppVersion(id int) (*InstalledAppVersions, error) {
	model := &InstalledAppVersions{}
	err := impl.dbConnection.Model(model).
		Column("installed_app_versions.*", "InstalledApp", "InstalledApp.App", "AppStoreApplicationVersion", "AppStoreApplicationVersion.AppStore").
		Column("AppStoreApplicationVersion.AppStore.ChartRepo").
		Where("installed_app_versions.id = ?", id).Where("installed_app_versions.active = true").Select()
	return model, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppVersionAny(id int) (*InstalledAppVersions, error) {
	model := &InstalledAppVersions{}
	err := impl.dbConnection.Model(model).
		Column("installed_app_versions.*", "InstalledApp", "InstalledApp.App", "AppStoreApplicationVersion", "AppStoreApplicationVersion.AppStore").
		Column("AppStoreApplicationVersion.AppStore.ChartRepo").
		Where("installed_app_versions.id = ?", id).Select()
	return model, err
}

func (impl InstalledAppRepositoryImpl) GetAllInstalledApps(filter *appStoreBean.AppStoreFilter) ([]InstalledAppsWithChartDetails, error) {
	var installedAppsWithChartDetails []InstalledAppsWithChartDetails
	var query string
	query = "select iav.updated_on, iav.id as installed_app_version_id, ch.name as chart_repo_name,"
	query = query + " env.environment_name, env.id as environment_id, a.app_name, asav.icon, asav.name as app_store_application_name,"
	query = query + " env.namespace, cluster.cluster_name, a.team_id, cluster.id as cluster_id, "
	query = query + " asav.id as app_store_application_version_id, ia.id , asav.deprecated"
	query = query + " from installed_app_versions iav"
	query = query + " inner join installed_apps ia on iav.installed_app_id = ia.id"
	query = query + " inner join app a on a.id = ia.app_id"
	query = query + " inner join environment env on ia.environment_id = env.id"
	query = query + " inner join cluster on env.cluster_id = cluster.id"
	query = query + " inner join app_store_application_version asav on iav.app_store_application_version_id = asav.id"
	query = query + " inner join app_store aps on aps.id = asav.app_store_id"
	query = query + " inner join chart_repo ch on ch.id = aps.chart_repo_id"
	query = query + " where ia.active = true and iav.active = true"
	if filter.OnlyDeprecated {
		query = query + " AND asav.deprecated = TRUE"
	}
	if len(filter.AppStoreName) > 0 {
		query = query + " AND aps.name LIKE '%" + filter.AppStoreName + "%'"
	}
	if len(filter.AppName) > 0 {
		query = query + " AND a.app_name LIKE '%" + filter.AppName + "%'"
	}
	if len(filter.ChartRepoId) > 0 {
		query = query + " AND ch.id IN (" + sqlIntSeq(filter.ChartRepoId) + ")"
	}
	if len(filter.EnvIds) > 0 {
		query = query + " AND env.id IN (" + sqlIntSeq(filter.EnvIds) + ")"
	}
	if len(filter.ClusterIds) > 0 {
		query = query + " AND cluster.id IN (" + sqlIntSeq(filter.ClusterIds) + ")"
	}
	query = query + " ORDER BY aps.name ASC"
	if filter.Size > 0 {
		query = query + " OFFSET " + strconv.Itoa(filter.Offset) + " LIMIT " + strconv.Itoa(filter.Size) + ""
	}
	query = query + ";"
	var err error
	_, err = impl.dbConnection.Query(&installedAppsWithChartDetails, query)
	if err != nil {
		return nil, err
	}
	return installedAppsWithChartDetails, err
}

func (impl InstalledAppRepositoryImpl) GetAllIntalledAppsByAppStoreId(appStoreId int) ([]InstalledAppAndEnvDetails, error) {
	var installedAppAndEnvDetails []InstalledAppAndEnvDetails
	var queryTemp = "select env.environment_name, env.id as environment_id, a.app_name, a.app_offering_mode, ia.updated_on, u.email_id, asav.id as app_store_application_version_id, iav.id as installed_app_version_id, ia.id as installed_app_id " +
		" from installed_app_versions iav inner join installed_apps ia on iav.installed_app_id = ia.id" +
		" inner join app a on a.id = ia.app_id " +
		" inner join app_store_application_version asav on iav.app_store_application_version_id = asav.id " +
		" inner join app_store aps on asav.app_store_id = aps.id " +
		" inner join environment env on ia.environment_id = env.id " +
		" left join users u on u.id = ia.updated_by " +
		" where aps.id = " + strconv.Itoa(appStoreId) + " and ia.active=true and iav.active=true and env.active=true"
	_, err := impl.dbConnection.Query(&installedAppAndEnvDetails, queryTemp)
	if err != nil {
		return nil, err
	}
	return installedAppAndEnvDetails, err
}

func (impl InstalledAppRepositoryImpl) GetAllInstalledAppsByChartRepoId(chartRepoId int) ([]InstalledAppAndEnvDetails, error) {
	var installedAppAndEnvDetails []InstalledAppAndEnvDetails
	var queryTemp = "select env.environment_name, env.id as environment_id, a.app_name, ia.updated_on, u.email_id, asav.id as app_store_application_version_id, iav.id as installed_app_version_id, ia.id as installed_app_id " +
		" from installed_app_versions iav inner join installed_apps ia on iav.installed_app_id = ia.id" +
		" inner join app a on a.id = ia.app_id " +
		" inner join app_store_application_version asav on iav.app_store_application_version_id = asav.id " +
		" inner join app_store aps on asav.app_store_id = aps.id " +
		" inner join environment env on ia.environment_id = env.id " +
		" left join users u on u.id = ia.updated_by " +
		" where aps.chart_repo_id = ? and ia.active=true and iav.active=true and env.active=true"
	_, err := impl.dbConnection.Query(&installedAppAndEnvDetails, queryTemp, chartRepoId)
	if err != nil {
		return nil, err
	}
	return installedAppAndEnvDetails, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppVersionByInstalledAppIdAndEnvId(installedAppId int, envId int) (*InstalledAppVersions, error) {
	installedAppVersion := &InstalledAppVersions{}
	err := impl.dbConnection.
		Model(installedAppVersion).
		Column("installed_app_versions.*", "InstalledApp", "InstalledApp.App", "InstalledApp.Environment", "AppStoreApplicationVersion", "AppStoreApplicationVersion.AppStore", "AppStoreApplicationVersion.AppStore.ChartRepo").
		Join("inner join installed_apps ia on ia.id = installed_app_versions.installed_app_id").
		Where("ia.id = ?", installedAppId).
		Where("ia.environment_id = ?", envId).
		Where("ia.active = true").Where("installed_app_versions.active = true").
		Limit(1).
		Select()
	return installedAppVersion, err
}

func sqlIntSeq(ns []int) string {
	if len(ns) == 0 {
		return ""
	}
	estimate := len(ns) * 4
	b := make([]byte, 0, estimate)
	for _, n := range ns {
		b = strconv.AppendInt(b, int64(n), 10)
		b = append(b, ',')
	}
	b = b[:len(b)-1]
	return string(b)
}

func (impl InstalledAppRepositoryImpl) DeleteInstalledApp(model *InstalledApps) (*InstalledApps, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl InstalledAppRepositoryImpl) DeleteInstalledAppVersion(model *InstalledAppVersions) (*InstalledAppVersions, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppVersionByInstalledAppId(id int) ([]*InstalledAppVersions, error) {
	model := make([]*InstalledAppVersions, 0)
	err := impl.dbConnection.Model(&model).
		Column("installed_app_versions.*").
		Where("installed_app_versions.installed_app_id = ?", id).
		Where("installed_app_versions.active = true").Select()

	return model, err
}

func (impl *InstalledAppRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return impl.dbConnection
}

func (impl InstalledAppRepositoryImpl) GetClusterComponentByClusterId(clusterId int) ([]*InstalledApps, error) {
	var models []*InstalledApps
	err := impl.dbConnection.Model(&models).
		Column("installed_apps.*", "App", "Environment").
		Where("environment.cluster_id = ?", clusterId).
		Where("installed_apps.active = ?", true).
		Where("environment.active = ?", true).
		Select()
	return models, err
}

func (impl InstalledAppRepositoryImpl) GetClusterComponentByClusterIds(clusterIds []int) ([]*InstalledApps, error) {
	var models []*InstalledApps
	err := impl.dbConnection.Model(&models).
		Column("installed_apps.*", "App", "Environment").
		Where("environment.cluster_id in (?)", pg.In(clusterIds)).
		Where("installed_apps.active = ?", true).
		Where("environment.active = ?", true).
		Select()
	return models, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppVersionByAppIdAndEnvId(appId int, envId int) (*InstalledAppVersions, error) {
	installedAppVersion := &InstalledAppVersions{}
	err := impl.dbConnection.
		Model(installedAppVersion).
		Column("installed_app_versions.*", "InstalledApp", "InstalledApp.App", "InstalledApp.Environment", "AppStoreApplicationVersion", "AppStoreApplicationVersion.AppStore", "AppStoreApplicationVersion.AppStore.ChartRepo").
		Join("inner join installed_apps ia on ia.id = installed_app_versions.installed_app_id").
		Where("ia.app_id = ?", appId).
		Where("ia.environment_id = ?", envId).
		Where("ia.active = true").Where("installed_app_versions.active = true").
		Order("installed_app_versions.id DESC").
		Limit(1).
		Select()
	return installedAppVersion, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppVersionByClusterIds(clusterIds []int) ([]*InstalledAppVersions, error) {
	var installedAppVersions []*InstalledAppVersions
	err := impl.dbConnection.
		Model(&installedAppVersions).
		Column("installed_app_versions.*", "InstalledApp", "InstalledApp.App", "InstalledApp.Environment", "AppStoreApplicationVersion", "AppStoreApplicationVersion.AppStore", "AppStoreApplicationVersion.AppStore.ChartRepo").
		Join("inner join installed_apps ia on ia.id = installed_app_versions.installed_app_id").
		Join("inner join environment env on env.id = ia.environment_id").
		Where("ia.active = true").Where("installed_app_versions.active = true").
		Where("env.cluster_id in (?)", pg.In(clusterIds)).Where("env.active = ?", true).
		Order("installed_app_versions.id desc").
		Select()
	return installedAppVersions, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppVersionByClusterIdsV2(clusterIds []int) ([]*InstalledAppVersions, error) {
	var installedAppVersions []*InstalledAppVersions
	err := impl.dbConnection.
		Model(&installedAppVersions).
		Column("installed_app_versions.*", "InstalledApp", "InstalledApp.App", "InstalledApp.Environment", "AppStoreApplicationVersion", "AppStoreApplicationVersion.AppStore", "AppStoreApplicationVersion.AppStore.ChartRepo").
		Join("inner join installed_apps ia on ia.id = installed_app_versions.installed_app_id").
		Join("inner join cluster_installed_apps cia on cia.installed_app_id = ia.id").
		Where("ia.active = true").Where("installed_app_versions.active = true").Where("cia.cluster_id in (?)", pg.In(clusterIds)).
		Order("installed_app_versions.id desc").
		Select()
	return installedAppVersions, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledApplicationByClusterIdAndNamespaceAndAppName(clusterId int, namespace string, appName string) (*InstalledApps, error) {
	model := &InstalledApps{}
	err := impl.dbConnection.Model(model).
		Column("installed_apps.*", "App", "Environment").
		Where("environment.cluster_id = ?", clusterId).
		Where("environment.namespace = ?", namespace).
		Where("app.app_name = ?", appName).
		Where("installed_apps.active = ?", true).
		Where("app.active = ?", true).
		Where("environment.active = ?", true).
		Select()
	return model, err
}