package manageusers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/foomo/htpasswd"
	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
)

// Config type for the handler
type Config struct {
	Route        string
	HtPasswdFile string
	UserInfoFile string
}

// User type
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Origin   string `json:"origin"`
	Role     string `json:"role"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
}

// UserInfos type
type UserInfos struct {
	Sub    string `json:"sub"`
	Origin string `json:"origin"`
	Claims Claims `json:"claims"`
}

// Claims struct
type Claims struct {
	Role    string `json:"role"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
}

func init() {
	caddy.RegisterPlugin("manageusers", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
	fmt.Printf("init manageusers\n")
}

func setup(c *caddy.Controller) error {
	// Default configuration
	config := Config{
		Route:        "/manageusers",
		HtPasswdFile: "./.htpasswd",
		UserInfoFile: "",
	}
	// Can be altered by args
	for c.Next() {
		if c.NextArg() {
			config.Route = c.Val() // The route to access user management
		}
		if c.NextArg() {
			config.HtPasswdFile = c.Val() // The .htpasswd file
		}
		if c.NextArg() {
			config.UserInfoFile = c.Val() // The user file
		}
	}

	cfg := httpserver.GetConfig(c)
	mid := func(next httpserver.Handler) httpserver.Handler {
		return PwdHandler{Next: next, Config: config}
	}
	cfg.AddMiddleware(mid)

	return nil
}

// PwdHandler type
type PwdHandler struct {
	Next   httpserver.Handler
	Config Config
}

func (h PwdHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if httpserver.Path(r.URL.Path).Matches(h.Config.Route) {
		w.Header().Set("Content-Type", "text/plain")
		htPasswdFile := h.Config.HtPasswdFile
		userInfoFile := h.Config.UserInfoFile

		switch r.Method {

		case "GET":
			// Get users list
			passwords, err := htpasswd.ParseHtpasswdFile(htPasswdFile)
			if err == nil {
				w.Header().Set("Content-Type", "application/json")
				users := make(map[string]User)
				for username := range passwords {
					users[username] = User{Username: username, Password: ""}
				}
				// if user file is set, join user info to response
				var usersFromJSON []UserInfos
				if userInfoFile != "" {
					usersFromJSON, err = ParseUserInfoFile(userInfoFile)
				}
				if err == nil {
					// do the actual merging
					mergeUsers(&users, usersFromJSON)
				}

				json.NewEncoder(w).Encode(users)
			} else {
				http.Error(w, "password file unreadable", 400)
			}

		case "POST":
			// Add an user or alter his password
			var u User
			if r.Body == nil {
				http.Error(w, "Please send a request body", 400)
			}
			jsonErr := json.NewDecoder(r.Body).Decode(&u)
			if jsonErr != nil {
				http.Error(w, jsonErr.Error(), 400)
			}
			pwdErr := htpasswd.SetPassword(htPasswdFile, u.Username, u.Password, htpasswd.HashBCrypt)
			var userInfoErr error
			if pwdErr == nil && userInfoFile != "" {
				userInfoErr = addUserToUserInfos(&u, userInfoFile)
			}
			if pwdErr == nil && userInfoErr == nil {
				fmt.Fprintf(w, "User %v added or altered", u.Username)
				doReload()
			} else {
				http.Error(w, fmt.Sprintf("Error adding or altering user %v", u.Username), 400)
			}

		case "DELETE":
			// Remove an user
			userToDelete := path.Base(r.URL.Path)
			// Concurrently delete user from both file
			errChan := make(chan error)
			go func() { errChan <- htpasswd.RemoveUser(htPasswdFile, userToDelete) }()
			var userInfoErr error
			if userInfoFile != "" {
				userInfoErr = deleteUserFromUserInfos(userToDelete, userInfoFile)
			}
			pwdErr := <-errChan
			if pwdErr == nil && userInfoErr == nil {
				fmt.Fprintf(w, "User %v deleted", userToDelete)
				doReload()
			} else {
				http.Error(w, fmt.Sprintf("Error deleting user %v", userToDelete), 400)
			}

		default:
			http.Error(w, "Method not allowed", 405)
		}

		return 0, nil

	}

	return h.Next.ServeHTTP(w, r)
}

func addUserToUserInfos(u *User, file string) error {
	// Read the existing user info
	userInfos, err := ParseUserInfoFile(file)
	if err != nil {
		return err
	}
	// Create the user info literal from the user
	newUserInfo := UserInfos{
		Sub:    u.Username,
		Origin: u.Origin,
		Claims: Claims{
			Role:    u.Role,
			Name:    u.Name,
			Surname: u.Surname,
		},
	}

	// Search for existing user
	found := false
	for i, userInfo := range userInfos {
		if userInfo.Sub == u.Username {
			// If found, update it
			userInfos[i] = newUserInfo
			found = true
			break
		}
	}
	// If not append it
	if !found {
		userInfos = append(userInfos, newUserInfo)
	}
	// Write the new file
	return WriteUserInfoFile(file, &userInfos)
}

func deleteUserFromUserInfos(u string, file string) error {
	// Read the existing user info
	userInfos, err := ParseUserInfoFile(file)
	if err != nil {
		return err
	}

	// Search for existing user
	for i, userInfo := range userInfos {
		if userInfo.Sub == u {
			// If found, delete it
			userInfos = append(userInfos[:i], userInfos[i+1:]...)
			break
		}
	}
	// Write the new file
	return WriteUserInfoFile(file, &userInfos)
}

// WriteUserInfoFile write user infos into file
func WriteUserInfoFile(file string, users *[]UserInfos) error {
	jsonData, err := json.Marshal(*users)

	if err != nil {
		return err
	}
	jsonFile, err := os.Create(file)

	if err != nil {
		return err
	}
	defer jsonFile.Close()

	jsonFile.Write(jsonData)
	jsonFile.Close()
	return nil
}

// ParseUserInfoFile load a json user file
func ParseUserInfoFile(file string) (users []UserInfos, err error) {
	userInfoFile, err := os.Open(file)
	defer userInfoFile.Close()
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	err = json.NewDecoder(userInfoFile).Decode(&users)
	return
}

// Merge user infos into user
func mergeUsers(As *map[string]User, Bs []UserInfos) error {
	if len(*As) == 0 {
		return errors.New("User map is empty")
	}
	bMap := make(map[string]UserInfos)
	for _, b := range Bs {
		bMap[b.Sub] = b
	}
	for key, a := range *As {
		if b, ok := bMap[a.Username]; ok {
			a.Origin = b.Origin
			a.Role = b.Claims.Role
			a.Name = b.Claims.Name
			a.Surname = b.Claims.Surname
			(*As)[key] = a
		}
	}
	return nil
}
