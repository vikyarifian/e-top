package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"etop/auth"
	"etop/db"
	"etop/models"
	"etop/services"
	"etop/templates/components/ui"
	"etop/templates/features"
	"etop/templates/layouts"
	"etop/templates/pages"
	"etop/utils"

	"gorm.io/gorm"
)

func HandleCreateDepartmentForm(w http.ResponseWriter, r *http.Request) error {
	return features.CreateDeptForm().Render(r.Context(), w)
}

// func HandleEditDepartmentForm(w http.ResponseWriter, r *http.Request) error {
// 	dept_id := r.URL.Query().Get("dept_id")
// 	dept := models.Department{}
// 	if err := db.PgSql.Where("id=?", dept_id).Preload("Members", func(db *gorm.DB) *gorm.DB {
// 		return db.Preload("User")
// 	}).Preload("Tags").First(&dept).Error; err != nil {
// 		w.WriteHeader(http.StatusNotFound)
// 		return pages.NotFound().Render(r.Context(), w)
// 	}
// 	return features.EditDeptForm(dept).Render(r.Context(), w)
// }

func HandleDepartments(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetAuth(w, r)
	switch r.Method {
	case http.MethodGet:
		id := r.URL.Query().Get("id")
		if strings.Trim(id, " ") != "" {
			var dept models.Department
			if err := db.PgSql.Where("id=?", id).
				Preload("Members", func(db *gorm.DB) *gorm.DB {
					return db.Preload("User")
				}).Order("no").First(&dept).Error; err != nil {
				w.WriteHeader(http.StatusNotFound)
				return layouts.Layout("404 Not Found", user, pages.NotFound()).Render(r.Context(), w)
			}
			authorized := false
			for _, member := range dept.Members {
				if member.UserID == user.ID {
					authorized = true
					break
				}
			}
			if !authorized {
				return layouts.Layout("403 Forbidden", user, pages.Forbidden()).Render(r.Context(), w)
			}
			return layouts.Layout("Department", user, features.Department(dept)).Render(r.Context(), w)
		}

		var depts []models.Department
		db.PgSql.
			Preload("Members", func(db *gorm.DB) *gorm.DB {
				return db.Preload("User")
			}).Order("no").Find(&depts)
		return layouts.Layout("Departments", user, features.Departments(depts, user)).Render(r.Context(), w)
	case http.MethodPost:
		id := r.URL.Query().Get("id")
		if strings.Trim(id, " ") != "" {
			var dept models.Department
			if err := db.PgSql.Where("id=?", id).
				Preload("Members", func(db *gorm.DB) *gorm.DB {
					return db.Preload("User")
				}).Order("no").First(&dept).Error; err != nil {
				w.WriteHeader(http.StatusNotFound)
				return pages.NotFound().Render(r.Context(), w)
			}
			authorized := false
			for _, member := range dept.Members {
				if member.UserID == user.ID {
					authorized = true

					break
				}
			}
			if !authorized {
				return layouts.Layout("403 Forbidden", user, pages.Forbidden()).Render(r.Context(), w)
			}

			return features.Department(dept).Render(r.Context(), w)
		}

		var depts []models.Department
		db.PgSql.Preload("Members").Order("no").Find(&depts)
		return features.Departments(depts, user).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}

func HandleDepartment(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetAuth(w, r)
	if user.Level != "ADMIN" {
		w.WriteHeader(http.StatusUnauthorized)
		return ui.Toast("dept-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
	} else {
		switch r.Method {
		case http.MethodPost:
			var dept models.Department
			r.ParseForm()
			dept.Name = r.FormValue("name")
			dept.Description = r.FormValue("description")
			dept.Color = r.FormValue("color")
			dept.DeptHeadID = r.FormValue("dept_head_id")

			if len(strings.Trim(dept.Name, " ")) < 1 {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("dept-error", "warning", "", "Name is required!", "", nil).Render(r.Context(), w)
			}

			if len(strings.Trim(dept.Color, " ")) < 1 {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("dept-error", "warning", "", "Color is required!", "", nil).Render(r.Context(), w)
			}

			if len(strings.Trim(dept.DeptHeadID, " ")) < 1 {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("dept-error", "warning", "", "Manager is required!", "", nil).Render(r.Context(), w)
			}

			no := 0
			user, _ := auth.GetAuth(w, r)
			t := time.Now()
			db.PgSql.Table("departments").Select("max(no)").Row().Scan(&no)
			dept.ID = utils.GenerateHash(strconv.Itoa(no + 1))
			dept.CreatedAt = &t
			dept.CreatedBy = user.ID
			dept.UpdatedAt = &t
			dept.UpdatedBy = user.ID

			if err := db.PgSql.Create(&dept).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return ui.Toast("dept-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
			}

			services.AddLog(user.ID, "created_dept", "Dept", dept.ID, map[string]any{
				"description": "created dept " + dept.Name,
			})

			members := []models.DepartmentMember{}
			member := models.DepartmentMember{}
			no = 0
			db.PgSql.Table("department_members").Select("max(no)").Row().Scan(&no)
			member.DepartmentID = dept.ID
			member.UserID = dept.DeptHeadID
			member.Role = "MANAGER"
			member.CreatedAt = &t
			member.CreatedBy = user.ID
			member.UpdatedAt = &t
			member.UpdatedBy = user.ID

			members = append(members, member)

			var membersPayload []string
			if err := json.Unmarshal([]byte(r.FormValue("members")), &membersPayload); err == nil {
				for _, m := range membersPayload {
					if m != dept.DeptHeadID {
						members = append(members, models.DepartmentMember{
							// ID:          utils.GenerateHash(strconv.Itoa(no + 2 + i)),
							DepartmentID: dept.ID,
							Role:         "MEMBER",
							UserID:       m,
							CreatedAt:    &t,
							CreatedBy:    user.ID,
							UpdatedAt:    &t,
							UpdatedBy:    user.ID,
						})
					}
				}
			}

			db.PgSql.Create(&members)
			return ui.Toast("dept-success", "success", "", "Department created successfully.", "", nil).Render(r.Context(), w)
		case http.MethodPut:
			var dept models.Department
			r.ParseForm()

			dept.ID = r.FormValue("dept_id")

			membersPayload := []models.WorkspaceMember{}
			if err := json.Unmarshal([]byte(r.FormValue("members")), &membersPayload); err == nil {
				for _, m := range membersPayload {
					var member = models.WorkspaceMember{}
					db.PgSql.Where("user_id=? and dept_id=?", m.UserID, dept.ID).First(&member)
					member.Role = m.Role
					db.PgSql.Save(&member)
				}

			}

			user, _ := auth.GetAuth(w, r)
			t := time.Now()
			if err := db.PgSql.Where("id=?", dept.ID).First(&dept).Error; err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("dept-error", "warning", "", "Department is not found!", "", nil).Render(r.Context(), w)
			}

			var member models.DepartmentMember
			if err := db.PgSql.Where("workspace_id=? AND user_id=?", dept.ID, user.ID).First(&member).Error; err != nil {
				println(err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("dept-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
			}

			updated := false
			logDetails := map[string]any{
				"description": "updated workspace",
			}

			if strings.Trim(dept.Name, " ") != strings.Trim(r.FormValue("name"), " ") {
				logDetails["old_name"] = dept.Name
				logDetails["new_name"] = r.FormValue("name")
				updated = true
			}
			dept.Name = r.FormValue("name")

			if strings.Trim(dept.Description, " ") != strings.Trim(r.FormValue("description"), " ") {
				logDetails["old_description"] = dept.Description
				logDetails["new_description"] = r.FormValue("description")
				updated = true
			}
			dept.Description = r.FormValue("description")

			if strings.Trim(dept.Color, " ") != strings.Trim(r.FormValue("color"), " ") {
				logDetails["old_color"] = dept.Color
				logDetails["new_color"] = r.FormValue("color")
				updated = true
			}
			dept.Color = r.FormValue("color")

			if strings.Trim(dept.DeptHeadID, " ") != strings.Trim(r.FormValue("dept_head_id"), " ") {
				logDetails["old_manager"] = dept.DeptHead
				logDetails["new_manager"] = r.FormValue("manager")
				updated = true
			}
			dept.DeptHeadID = r.FormValue("dept_head_id")

			if len(strings.Trim(dept.Name, " ")) < 1 {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("dept-error", "warning", "", "Name is required!", "", nil).Render(r.Context(), w)
			}

			if len(strings.Trim(dept.Color, " ")) < 1 {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("dept-error", "warning", "", "Color is required!", "", nil).Render(r.Context(), w)
			}

			if len(strings.Trim(dept.DeptHeadID, " ")) < 1 {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("dept-error", "warning", "", "Manager is required!", "", nil).Render(r.Context(), w)
			}

			if strings.ToUpper(strings.Trim(member.Role, " ")) != "ADMIN" && strings.ToUpper(strings.Trim(member.Role, " ")) != "OWNER" {
				w.WriteHeader(http.StatusBadRequest)
				return ui.Toast("workspace-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
			}

			dept.UpdatedAt = &t
			dept.UpdatedBy = user.ID

			if err := db.PgSql.Save(&dept).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return ui.Toast("dept-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
			}

			if updated {
				services.AddLog(user.ID, "updated_dept", "Dept", dept.ID, logDetails)
			}

			return ui.Toast("dept-success", "success", "", "Department updated successfully.", "", nil).Render(r.Context(), w)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return nil
		}
	}
}
