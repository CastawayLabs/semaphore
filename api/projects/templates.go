package projects

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/castawaylabs/mulekick"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
)

func TemplatesMiddleware(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	templateID, err := util.GetIntParam("template_id", w, r)
	if err != nil {
		return
	}

	var template db.Template
	if err := db.Mysql.SelectOne(&template, "select * from project__template where project_id=? and id=?", project.ID, templateID); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		panic(err)
	}

	context.Set(r, "template", template)
}

func GetTemplates(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	var templates []db.Template

	q := squirrel.Select("*").
		From("project__template").
		Where("project_id=?", project.ID)

	query, args, _ := q.ToSql()

	if _, err := db.Mysql.Select(&templates, query, args...); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusOK, templates)
}

func AddTemplate(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)

	var template db.Template
	if err := mulekick.Bind(w, r, &template); err != nil {
		return
	}

	res, err := db.Mysql.Exec("insert into project__template set ssh_key_id=?, project_id=?, inventory_id=?, repository_id=?, environment_id=?, alias=?, playbook=?, arguments=?, override_args=?", template.SshKeyID, project.ID, template.InventoryID, template.RepositoryID, template.EnvironmentID, template.Alias, template.Playbook, template.Arguments, template.OverrideArguments)
	if err != nil {
		panic(err)
	}

	insertID, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}

	template.ID = int(insertID)

	objType := "template"
	desc := "Template ID " + strconv.Itoa(template.ID) + " created"
	if err := (db.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &template.ID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	mulekick.WriteJSON(w, http.StatusCreated, template)
}

func UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	oldTemplate := context.Get(r, "template").(db.Template)

	var template db.Template
	if err := mulekick.Bind(w, r, &template); err != nil {
		return
	}

	if *template.Arguments == "" {
		template.Arguments = nil
	}

	if _, err := db.Mysql.Exec("update project__template set ssh_key_id=?, inventory_id=?, repository_id=?, environment_id=?, alias=?, playbook=?, arguments=?, override_args=? where id=?", template.SshKeyID, template.InventoryID, template.RepositoryID, template.EnvironmentID, template.Alias, template.Playbook, template.Arguments, template.OverrideArguments, oldTemplate.ID); err != nil {
		panic(err)
	}

	desc := "Template ID " + strconv.Itoa(template.ID) + " updated"
	objType := "template"
	if err := (db.Event{
		ProjectID:   &oldTemplate.ProjectID,
		Description: &desc,
		ObjectID:    &oldTemplate.ID,
		ObjectType:  &objType,
	}.Insert()); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func RemoveTemplate(w http.ResponseWriter, r *http.Request) {
	tpl := context.Get(r, "template").(db.Template)

	if _, err := db.Mysql.Exec("delete from project__template where id=?", tpl.ID); err != nil {
		panic(err)
	}

	desc := "Template ID " + strconv.Itoa(tpl.ID) + " deleted"
	if err := (db.Event{
		ProjectID:   &tpl.ProjectID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
