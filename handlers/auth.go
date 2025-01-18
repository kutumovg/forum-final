package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"forum/models"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
var (
	GoogleConfig = &oauth2.Config{
		ClientID:     ${{ secrets.GOOGLE_ID }},
		ClientSecret: ${{ secrets.GOOGLE_SECRET }},
		RedirectURL:  "https://forum.qarjy.kz/auth/google/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	GitHubConfig = &oauth2.Config{
		ClientID:     ${{ secrets.GHUB_ID }},
		ClientSecret: ${{ secrets.GHUB_SECRET }},
		RedirectURL:  "https://forum.qarjy.kz/auth/github/callback",
		Scopes:       []string{"read:user", "user:email"},
		Endpoint:     github.Endpoint,
	}
)

// authorization
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Validate email format
		if !isValidEmail(email) {
			ErrorHandler(w, r, http.StatusBadRequest, "Invalid email format")
			return
		}

		// Check if the email is already in use
		emailExists, err := models.CheckEmailExists(email)

		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
		//if in use
		if emailExists {
			tmpl, _ := template.ParseFiles("templates/register.html")
			tmpl.Execute(w, struct{ Error string }{Error: "Email is already registered"})
			return
		}

		// Check if the username is already in use
		usernameExists, err := models.CheckUsernameExists(username)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
		if usernameExists {
			tmpl, _ := template.ParseFiles("templates/register.html")
			tmpl.Execute(w, struct{ Error string }{Error: "Username is already taken"})
			return
		}

		// Register the user (create user in the database)
		sessionToken, err := models.RegisterUser(email, username, password)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}

		// Automatically log in the user after registration
		cookie := http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: time.Now().Add(24 * time.Hour),
		}
		http.SetCookie(w, &cookie)

		// Redirect to the main page after successful registration and login
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl, _ := template.ParseFiles("templates/register.html")
	tmpl.Execute(w, nil)
}

// LoginHandler - Handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Authenticate the user
		sessionToken, err := models.AuthenticateUser(email, password)
		if err != nil {
			tmpl, _ := template.ParseFiles("templates/login.html")
			tmpl.Execute(w, struct{ Error string }{Error: "Invalid email or password"})
			return
		}

		// Set session cookie
		cookie := http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: time.Now().Add(24 * time.Hour),
		}
		http.SetCookie(w, &cookie)

		// Redirect to the main page
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Render the login page
	tmpl, _ := template.ParseFiles("templates/login.html")
	tmpl.Execute(w, nil)
}

// LogoutHandler - Logs the user out by clearing the session cookie
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie
	cookie := http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour), // Expire the cookie immediately
	}
	http.SetCookie(w, &cookie)

	// Redirect to the main page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	url := GoogleConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token, err := GoogleConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var userInfo map[string]string
	json.Unmarshal(data, &userInfo)

	email := userInfo["email"]
	id := userInfo["id"]

	// Call AuthenticateOrCreateUser
	sessionToken, err := models.GoogleGithubUser(id, email, "google")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	cookie := http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Path:    "/", // Set cookie path
		Expires: time.Now().Add(24 * time.Hour),
	}
	http.SetCookie(w, &cookie)

	// Redirect to the main page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func GitHubLoginHandler(w http.ResponseWriter, r *http.Request) {
	url := GitHubConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GitHubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	token, err := GitHubConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	// resp, err := http.Get("https://api.github.com/user?access_token=" + token.AccessToken)
	// if err != nil {
	// 	http.Error(w, "Failed to get user info", http.StatusInternalServerError)
	// 	return
	// }
	// defer resp.Body.Close()

	// data, _ := io.ReadAll(resp.Body)
	// var userInfo map[string]interface{}
	// json.Unmarshal(data, &userInfo)

	// // email := userInfo["email"].(string)
	// email := fmt.Sprintf("%v", userInfo["email"])
	// id := fmt.Sprintf("%v", userInfo["id"]) // GitHub's ID might not be a string

	// Fetch user information from GitHub
	client := http.Client{}
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("GitHub API error: %v", resp.Status)
		http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}

	// Parse user information
	var userInfo struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
	}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	 if userInfo.Email == "" {
		 // Fallback: Get primary email address if not public
		 userInfo.Email, err = fetchPrimaryEmail(token.AccessToken)
		 if err != nil {
			 http.Error(w, "Failed to fetch primary email", http.StatusInternalServerError)
			 return
		 }
	 }

	// Call AuthenticateOrCreateUser
	sessionToken, err := models.GoogleGithubUser(fmt.Sprintf("%d", userInfo.ID), userInfo.Email, "github")
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	cookie := http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Path:    "/", // Set cookie path
		Expires: time.Now().Add(24 * time.Hour),
	}
	http.SetCookie(w, &cookie)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func fetchPrimaryEmail(accessToken string) (string, error) {
    client := http.Client{}
    req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
    if err != nil {
        return "", err
    }
    req.Header.Set("Authorization", "Bearer "+accessToken)

    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("GitHub API error: %v", resp.Status)
    }

    var emails []struct {
        Email   string `json:"email"`
        Primary bool   `json:"primary"`
        Verified bool  `json:"verified"`
    }
    err = json.NewDecoder(resp.Body).Decode(&emails)
    if err != nil {
        return "", err
    }

    for _, e := range emails {
        if e.Primary && e.Verified {
            return e.Email, nil
        }
    }

    return "", fmt.Errorf("no verified primary email found")
}
