package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"etop/auth"
	"etop/db"
	"etop/dto"
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

func HandleEditDepartmentForm(w http.ResponseWriter, r *http.Request) error {
	dept_id := r.URL.Query().Get("dept_id")
	var dept models.Department
	if err := db.PgSql.Where("id=?", dept_id).
		Preload("Members", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User")
		}).
		Preload("DeptHead").
		First(&dept).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		return ui.Toast("dept-error", "warning", "", "Department not found!", "", nil).Render(r.Context(), w)
	}
	return features.EditDeptForm(dept).Render(r.Context(), w)
}

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
				}).Preload("DeptHead").Order("no").First(&dept).Error; err != nil {
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
			return layouts.Layout("Department", user, features.Department(dept, user)).Render(r.Context(), w)
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
				}).Preload("DeptHead").Order("no").First(&dept).Error; err != nil {
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

			return features.Department(dept, user).Render(r.Context(), w)
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
	switch r.Method {
	case http.MethodGet:
		return handleDepartmentView(w, r, user, false)
	case http.MethodPost:
		id := r.URL.Query().Get("id")
		if strings.Trim(id, " ") != "" {
			return handleDepartmentView(w, r, user, true)
		}
		r.ParseForm()
		if r.FormValue("name") != "" {
			return handleDepartmentCreate(w, r, user)
		}
		return handleDepartmentView(w, r, user, true)
	case http.MethodPut:
		return handleDepartmentUpdate(w, r, user)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}

func handleDepartmentView(w http.ResponseWriter, r *http.Request, user dto.UserAuth, fragment bool) error {
	id := r.URL.Query().Get("id")
	if strings.Trim(id, " ") == "" {
		var myDept models.Department
		err := db.PgSql.
			Joins("JOIN department_members ON department_members.department_id = departments.id").
			Where("department_members.user_id = ?", user.ID).
			Preload("Members", func(db *gorm.DB) *gorm.DB {
				return db.Preload("User")
			}).
			Preload("DeptHead").
			First(&myDept).Error
		if err == nil {
			if fragment {
				return features.Department(myDept, user).Render(r.Context(), w)
			}
			return layouts.Layout(myDept.Name, user, features.Department(myDept, user)).Render(r.Context(), w)
		}
		var depts []models.Department
		db.PgSql.
			Preload("Members", func(db *gorm.DB) *gorm.DB {
				return db.Preload("User")
			}).
			Preload("DeptHead").
			Order("no").Find(&depts)
		if fragment {
			return features.DepartmentsJoin(depts, user).Render(r.Context(), w)
		}
		return layouts.Layout("Departments", user, features.DepartmentsJoin(depts, user)).Render(r.Context(), w)
	}

	var dept models.Department
	if err := db.PgSql.Where("id=?", id).
		Preload("Members", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User")
		}).Preload("DeptHead").Order("no").First(&dept).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		if fragment {
			return pages.NotFound().Render(r.Context(), w)
		}
		return layouts.Layout("404 Not Found", user, pages.NotFound()).Render(r.Context(), w)
	}

	isMember := false
	for _, m := range dept.Members {
		if m.UserID == user.ID {
			isMember = true
			break
		}
	}

	if isMember {
		if fragment {
			return features.Department(dept, user).Render(r.Context(), w)
		}
		return layouts.Layout(dept.Name, user, features.Department(dept, user)).Render(r.Context(), w)
	}

	if fragment {
		return features.DepartmentJoin(dept, user).Render(r.Context(), w)
	}
	return layouts.Layout(dept.Name, user, features.DepartmentJoin(dept, user)).Render(r.Context(), w)
}

func handleDepartmentCreate(w http.ResponseWriter, r *http.Request, user dto.UserAuth) error {
	if user.Level != "ADMIN" {
		w.WriteHeader(http.StatusUnauthorized)
		return ui.Toast("dept-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
	}

	var dept models.Department
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
}

func handleDepartmentUpdate(w http.ResponseWriter, r *http.Request, user dto.UserAuth) error {
	if user.Level != "ADMIN" {
		w.WriteHeader(http.StatusUnauthorized)
		return ui.Toast("dept-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
	}

	var dept models.Department
	r.ParseForm()
	dept.ID = r.FormValue("dept_id")

	t := time.Now()
	if err := db.PgSql.Where("id=?", dept.ID).First(&dept).Error; err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return ui.Toast("dept-error", "warning", "", "Department is not found!", "", nil).Render(r.Context(), w)
	}

	membersPayload := []struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}{}
	if err := json.Unmarshal([]byte(r.FormValue("members")), &membersPayload); err == nil {
		for _, m := range membersPayload {
			var member models.DepartmentMember
			db.PgSql.Where("user_id=? AND department_id=?", m.UserID, dept.ID).First(&member)
			member.Role = m.Role
			member.DepartmentID = dept.ID
			db.PgSql.Save(&member)
		}
	}

	updated := false
	logDetails := map[string]any{
		"description": "updated department",
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
		logDetails["old_manager"] = dept.DeptHeadID
		logDetails["new_manager"] = r.FormValue("dept_head_id")
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
}

func HandleInviteDepartmentForm(w http.ResponseWriter, r *http.Request) error {
	deptID := r.URL.Query().Get("dept_id")

	var dept models.Department
	if err := db.PgSql.Where("id=?", deptID).First(&dept).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		return ui.Toast("dept-error", "warning", "", "Department not found!", "", nil).Render(r.Context(), w)
	}

	users := []models.User{}
	db.PgSql.Where("id NOT IN (SELECT user_id FROM department_members WHERE department_id=?) AND id NOT IN (SELECT dept_head_id FROM departments WHERE id=?)", deptID, deptID).
		Find(&users)

	userRole := []models.UserRole{}
	for _, u := range users {
		var m models.UserRole
		m.ID = u.ID
		m.FullName = u.FullName
		m.Color = u.Color
		userRole = append(userRole, m)
	}

	return features.InviteDepartmentForm(deptID, dept.Name, dept.Color, userRole).Render(r.Context(), w)
}

func HandleInviteDepartment(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetAuth(w, r)
	switch r.Method {
	case http.MethodPost:
		r.ParseForm()
		deptID := r.FormValue("dept_id")

		var dept models.Department
		if err := db.PgSql.Where("id=?", deptID).First(&dept).Error; err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("dept-error", "warning", "", "Department is not found!", "", nil).Render(r.Context(), w)
		}

		authorized := false
		var myMember models.DepartmentMember
		if err := db.PgSql.Where("department_id=? AND user_id=?", deptID, user.ID).First(&myMember).Error; err == nil {
			if myMember.Role == "MANAGER" || user.Level == "ADMIN" {
				authorized = true
			}
		}
		if !authorized {
			w.WriteHeader(http.StatusUnauthorized)
			return ui.Toast("dept-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
		}

		var inviteMembers []string
		if err := json.Unmarshal([]byte(r.FormValue("invite_members")), &inviteMembers); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("dept-error", "warning", "", "Invalid members data!", "", nil).Render(r.Context(), w)
		}

		t := time.Now()
		no := 0
		db.PgSql.Table("department_members").Select("max(no)").Row().Scan(&no)
		members := []models.DepartmentMember{}
		for _, m := range inviteMembers {
			no++
			members = append(members, models.DepartmentMember{
				DepartmentID: deptID,
				UserID:       m,
				Role:         "MEMBER",
				CreatedAt:    &t,
				CreatedBy:    user.ID,
				UpdatedAt:    &t,
				UpdatedBy:    user.ID,
			})
		}

		if err := db.PgSql.Create(&members).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("dept-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
		}

		services.AddLog(user.ID, "invited_dept_members", "Dept", deptID, map[string]any{
			"description": "invited members to " + dept.Name,
		})

		return ui.Toast("dept-success", "success", "", "Members invited successfully.", "", nil).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}

func HandleJoinDepartment(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetAuth(w, r)
	switch r.Method {
	case http.MethodPost:
		deptID := r.URL.Query().Get("id")

		var dept models.Department
		if err := db.PgSql.Where("id=?", deptID).First(&dept).Error; err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return ui.Toast("dept-error", "warning", "", "Department is not found!", "", nil).Render(r.Context(), w)
		}

		var count int64
		db.PgSql.Where("department_id=? AND user_id=?", deptID, user.ID).Count(&count)
		if count > 0 {
			return ui.Toast("dept-error", "warning", "", "You're already a member of this department!", "", nil).Render(r.Context(), w)
		}

		t := time.Now()
		no := 0
		db.PgSql.Table("department_members").Select("max(no)").Row().Scan(&no)
		member := models.DepartmentMember{
			DepartmentID: deptID,
			UserID:       user.ID,
			Role:         "MEMBER",
			CreatedAt:    &t,
			CreatedBy:    user.ID,
			UpdatedAt:    &t,
			UpdatedBy:    user.ID,
		}

		if err := db.PgSql.Create(&member).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return ui.Toast("dept-error", "error", "", err.Error(), "", nil).Render(r.Context(), w)
		}

		services.AddLog(user.ID, "joined_dept", "Dept", deptID, map[string]any{
			"description": user.FullName + " joined " + dept.Name,
		})

		return ui.Toast("dept-success", "success", "", "You've joined "+dept.Name+" successfully!", "", nil).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}
