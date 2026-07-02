package backend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"gorm.io/gorm"
)

type TerminalRequest struct {
	Command string `json:"command"`
	Args    string `json:"args"`
	Path    string `json:"path"`
}

type TerminalResponse struct {
	Output        string `json:"output"`
	IsHtml        bool   `json:"isHtml"`
	Path          string `json:"path"`
	TriggerCamera bool   `json:"triggerCamera"`
	TriggerGlitch bool   `json:"triggerGlitch"`
}

type VFSNode struct {
	IsDir       bool
	IsSecretCam bool
	IsPasswords bool
	Content     string
}

func getVFSNode(role string, username string, p string) (*VFSNode, bool) {
	p = path.Clean(p)

	// Admin paths
	if p == "/home/admin" || p == "/home/admin/system_root" || p == "/home/admin/system_root/config" || p == "/home/admin/system_root/users" {
		if role != "admin" {
			return nil, false
		}
		return &VFSNode{IsDir: true}, true
	}
	if p == "/home/admin/system_root/config/system_config.dat" {
		if role != "admin" {
			return nil, false
		}
		return &VFSNode{IsDir: false, Content: "SYS_MODE=PROD<br>FIREWALL=ACTIVE<br>AUTHOR=FARID"}, true
	}
	if p == "/home/admin/system_root/users/passwords.txt" {
		if role != "admin" {
			return nil, false
		}
		return &VFSNode{IsDir: false, IsPasswords: true}, true
	}
	if p == "/home/admin/system_root/secret_cam.sh" {
		if role != "admin" {
			return nil, false
		}
		return &VFSNode{IsDir: false, IsSecretCam: true, Content: "Initializing secure webcam uplink..."}, true
	}
	if p == "/home/admin/system_root/root_flag.txt" {
		if role != "admin" {
			return nil, false
		}
		return &VFSNode{IsDir: false, Content: "CTF{y0u_4r3_7h3_sys4dm1n_m4st3r}"}, true
	}

	// Shared paths
	if p == "/" || p == "/home" {
		return &VFSNode{IsDir: true}, true
	}

	// Dynamic Guest paths
	guestHome := fmt.Sprintf("/home/%s", username)
	guestSandbox := fmt.Sprintf("/home/%s/guest_sandbox", username)

	if p == guestHome || p == guestSandbox {
		return &VFSNode{IsDir: true}, true
	}

	if p == guestSandbox+"/readme.txt" {
		return &VFSNode{IsDir: false, Content: "Hello Guest! To gain more access, you need to find the hidden CTF flag. Try looking around. Type 'help' to see available commands."}, true
	}

	if p == guestSandbox+"/hint.txt" {
		return &VFSNode{IsDir: false, Content: "Hint: Have you tried searching the website source code or database tables? Also checkout system logs monitor! Enter the GTA San Andreas cheat code 'hesoyam' for a surprise."}, true
	}

	return nil, false
}

func listDirectory(role string, username string, p string) (string, bool) {
	p = path.Clean(p)
	node, exists := getVFSNode(role, username, p)
	if !exists || !node.IsDir {
		return "", false
	}

	var items []string

	dirStyle := func(name string) string {
		return fmt.Sprintf("<span style=\"color:#00ffff\">%s/</span>", name)
	}
	fileStyle := func(name string) string {
		return name
	}

	guestHome := fmt.Sprintf("/home/%s", username)
	guestSandbox := fmt.Sprintf("/home/%s/guest_sandbox", username)

	switch p {
	case "/":
		items = append(items, dirStyle("home"))
	case "/home":
		items = append(items, dirStyle(username))
		if role == "admin" {
			items = append(items, dirStyle("admin"))
		}
	case guestHome:
		items = append(items, dirStyle("guest_sandbox"))
	case guestSandbox:
		items = append(items, fileStyle("readme.txt"), fileStyle("hint.txt"))
	case "/home/admin":
		if role == "admin" {
			items = append(items, dirStyle("system_root"))
		}
	case "/home/admin/system_root":
		if role == "admin" {
			items = append(items, dirStyle("config"), dirStyle("users"), fileStyle("passwords.txt"), fileStyle("secret_cam.sh"), fileStyle("root_flag.txt"))
		}
	case "/home/admin/system_root/config":
		if role == "admin" {
			items = append(items, fileStyle("system_config.dat"))
		}
	case "/home/admin/system_root/users":
		if role == "admin" {
			// empty dir
		}
	}

	if len(items) == 0 {
		return "total 0", true
	}

	return strings.Join(items, " &nbsp;&nbsp; "), true
}

func HandleTerminalExecute(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, err := verifyJWT(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req TerminalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := strings.ToLower(strings.TrimSpace(req.Command))
	args := strings.TrimSpace(req.Args)

	// Clean initial path
	cleanPath := req.Path
	if cleanPath == "" {
		if claims.Role == "admin" {
			cleanPath = "/home/admin/system_root"
		} else {
			cleanPath = fmt.Sprintf("/home/%s/guest_sandbox", claims.Username)
		}
	}
	cleanPath = path.Clean(cleanPath)
	if !strings.HasPrefix(cleanPath, "/") {
		cleanPath = "/" + cleanPath
	}

	var res TerminalResponse
	res.IsHtml = false
	res.Path = cleanPath // Default to keep same path

	switch cmd {
	case "pwd":
		res.Output = cleanPath
	case "whoami":
		res.Output = fmt.Sprintf("CURRENT ENTITY: %s | ROLE: %s", claims.Username, claims.Role)
	case "ls":
		output, ok := listDirectory(claims.Role, claims.Username, cleanPath)
		if !ok {
			res.Output = "ls: Cannot access directory"
		} else {
			res.Output = output
			res.IsHtml = true
		}
	case "cd":
		if args == "" || args == "~" {
			if claims.Role == "admin" {
				res.Path = "/home/admin/system_root"
			} else {
				res.Path = fmt.Sprintf("/home/%s/guest_sandbox", claims.Username)
			}
			res.Output = ""
		} else {
			var targetPath string
			if strings.HasPrefix(args, "/") {
				targetPath = path.Clean(args)
			} else {
				targetPath = path.Clean(path.Join(cleanPath, args))
			}

			node, exists := getVFSNode(claims.Role, claims.Username, targetPath)
			if !exists {
				if strings.HasPrefix(targetPath, "/home/admin") && claims.Role != "admin" {
					res.Output = fmt.Sprintf("cd: %s: Permission denied", args)
				} else {
					res.Output = fmt.Sprintf("cd: %s: No such file or directory", args)
				}
				res.Path = cleanPath
			} else if !node.IsDir {
				res.Output = fmt.Sprintf("cd: %s: Not a directory", args)
				res.Path = cleanPath
			} else {
				res.Path = targetPath
				res.Output = ""
			}
		}
	case "cat":
		if args == "" {
			res.Output = "cat: missing operand"
		} else {
			var targetPath string
			if strings.HasPrefix(args, "/") {
				targetPath = path.Clean(args)
			} else {
				targetPath = path.Clean(path.Join(cleanPath, args))
			}

			node, exists := getVFSNode(claims.Role, claims.Username, targetPath)
			if !exists {
				if strings.HasPrefix(targetPath, "/home/admin") && claims.Role != "admin" {
					res.Output = fmt.Sprintf("cat: %s: Permission denied", args)
				} else {
					res.Output = fmt.Sprintf("cat: %s: No such file or directory", args)
				}
			} else if node.IsDir {
				res.Output = fmt.Sprintf("cat: %s: Is a directory", args)
			} else {
				if node.IsPasswords {
					var users []User
					if err := db.Find(&users).Error; err != nil {
						res.Output = "Error querying user database"
					} else {
						var out strings.Builder
						out.WriteString("DECRYPTED DATABASE ACCOUNTS:<br>")
						for _, u := range users {
							out.WriteString(fmt.Sprintf("USER: <span style=\"color:var(--accent-color)\">%s</span> | ROLE: %s | PASS: %s<br>", u.Username, u.Role, u.Password))
						}
						res.Output = out.String()
						res.IsHtml = true
					}
				} else if node.IsSecretCam {
					res.Output = node.Content
					res.TriggerCamera = true
				} else {
					res.Output = node.Content
				}
			}
		}
	case "secret_cam.sh", "./secret_cam.sh":
		if claims.Role != "admin" {
			res.Output = "Permission denied"
		} else {
			res.Output = "Initializing secure webcam uplink..."
			res.TriggerCamera = true
		}
	case "submit_flag":
		if args == "" {
			res.Output = "submit_flag: missing flag string"
		} else if args == "CTF{h4ck3r_m4st3r_f4r1d_2024}" || args == "CTF{y0u_4r3_7h3_sys4dm1n_m4st3r}" {
			var score CTFScore
			err := db.Where("username = ?", claims.Username).First(&score).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					db.Create(&CTFScore{Username: claims.Username, Score: 100})
					res.Output = `<span style="color:#22c55e">FLAG ACCEPTED! +100 POINTS. ACHIEVEMENT UNLOCKED.</span>`
					res.IsHtml = true
				} else {
					res.Output = "Database error"
				}
			} else {
				res.Output = `<span style="color:#eab308">FLAG ALREADY SUBMITTED.</span>`
				res.IsHtml = true
			}
		} else {
			res.Output = `<span style="color:red">INCORRECT FLAG. INCIDENT LOGGED.</span>`
			res.IsHtml = true
		}
	case "rm", "del", "drop":
		broadcastLog(fmt.Sprintf("[ALERT] User %s attempted dangerous command: rm %s", claims.Username, args))
		res.Output = `<span style="color:red">ERROR: PERMISSION DENIED. THIS INCIDENT WILL BE REPORTED.</span>`
		res.IsHtml = true
		res.TriggerGlitch = true
	case "db":
		if claims.Role != "admin" {
			res.Output = `<span style="color:red">ERROR: Access denied. Database queries are restricted to Administrator accounts.</span>`
			res.IsHtml = true
		} else {
			res.Output, res.IsHtml = executeDBCommand(args)
		}
	default:
		res.Output = fmt.Sprintf("COMMAND NOT FOUND ON SERVER: %s", cmd)
	}

	broadcastLog(fmt.Sprintf("[CMD] %s ran: %s %s", claims.Username, cmd, args))

	json.NewEncoder(w).Encode(res)
}

func executeDBCommand(args string) (string, bool) {
	subcmd := strings.ToLower(strings.TrimSpace(args))
	if subcmd == "" {
		return `Usage: db [tables|users|articles|contacts|projects|skills|analytics|leaderboard|logins]`, false
	}

	switch subcmd {
	case "tables":
		var tables []string
		err := db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE' ORDER BY table_name").Scan(&tables).Error
		if err != nil {
			return fmt.Sprintf("Error querying tables: %v", err), false
		}
		
		var b strings.Builder
		b.WriteString("<span style=\"color:var(--accent-color)\">POSTGRESQL TABLES IN PORTFOLIO_DB:</span><br>")
		b.WriteString("<table style=\"width:100%; border-collapse:collapse; margin-top:5px; font-family:'Fira Code', monospace; font-size:0.8rem; border: 1px solid var(--text-primary);\">")
		b.WriteString("<tr style=\"border-bottom: 2px solid var(--text-primary); background: var(--btn-bg);\">")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">TABLE_NAME</th>")
		b.WriteString("</tr>")
		for _, t := range tables {
			b.WriteString("<tr style=\"border-bottom: 1px solid var(--text-secondary);\">")
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", t))
			b.WriteString("</tr>")
		}
		b.WriteString("</table>")
		return b.String(), true

	case "users":
		var users []User
		err := db.Limit(50).Order("created_at desc").Find(&users).Error
		if err != nil {
			return fmt.Sprintf("Error querying users: %v", err), false
		}
		var b strings.Builder
		b.WriteString("<span style=\"color:var(--accent-color)\">RECENT USERS (MAX 50):</span><br>")
		b.WriteString("<table style=\"width:100%; border-collapse:collapse; margin-top:5px; font-family:'Fira Code', monospace; font-size:0.8rem; border: 1px solid var(--text-primary);\">")
		b.WriteString("<tr style=\"border-bottom: 2px solid var(--text-primary); background: var(--btn-bg);\">")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">ID</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">USERNAME</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">EMAIL</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">ROLE</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">CREATED_AT</th>")
		b.WriteString("</tr>")
		for _, u := range users {
			b.WriteString("<tr style=\"border-bottom: 1px solid var(--text-secondary);\">")
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%d</td>", u.ID))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(u.Username)))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(u.Email)))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(u.Role)))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", u.CreatedAt.Format("2006-01-02 15:04:05")))
			b.WriteString("</tr>")
		}
		b.WriteString("</table>")
		return b.String(), true

	case "articles":
		var articles []Article
		err := db.Find(&articles).Error
		if err != nil {
			return fmt.Sprintf("Error querying articles: %v", err), false
		}
		var b strings.Builder
		b.WriteString("<span style=\"color:var(--accent-color)\">SYSTEM ARTICLES:</span><br>")
		b.WriteString("<table style=\"width:100%; border-collapse:collapse; margin-top:5px; font-family:'Fira Code', monospace; font-size:0.8rem; border: 1px solid var(--text-primary);\">")
		b.WriteString("<tr style=\"border-bottom: 2px solid var(--text-primary); background: var(--btn-bg);\">")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">ID</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">TITLE</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">AUTHOR</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">CONTENT (TRUNCATED)</th>")
		b.WriteString("</tr>")
		for _, a := range articles {
			contentTrunc := a.Content
			if len(contentTrunc) > 50 {
				contentTrunc = contentTrunc[:47] + "..."
			}
			b.WriteString("<tr style=\"border-bottom: 1px solid var(--text-secondary);\">")
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(a.ID)))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(a.Title)))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(a.Author)))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(contentTrunc)))
			b.WriteString("</tr>")
		}
		b.WriteString("</table>")
		return b.String(), true

	case "contacts":
		var contacts []Contact
		err := db.Order("created_at desc").Find(&contacts).Error
		if err != nil {
			return fmt.Sprintf("Error querying contacts: %v", err), false
		}
		var b strings.Builder
		b.WriteString("<span style=\"color:var(--accent-color)\">LATEST CONTACT MESSAGES:</span><br>")
		b.WriteString("<table style=\"width:100%; border-collapse:collapse; margin-top:5px; font-family:'Fira Code', monospace; font-size:0.8rem; border: 1px solid var(--text-primary);\">")
		b.WriteString("<tr style=\"border-bottom: 2px solid var(--text-primary); background: var(--btn-bg);\">")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">ID</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">NAME</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">EMAIL</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">PHONE</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">MESSAGE</th>")
		b.WriteString("</tr>")
		for _, c := range contacts {
			b.WriteString("<tr style=\"border-bottom: 1px solid var(--text-secondary);\">")
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%d</td>", c.ID))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(c.Name)))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(c.Email)))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(c.Phone)))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(c.Message)))
			b.WriteString("</tr>")
		}
		b.WriteString("</table>")
		return b.String(), true

	case "projects":
		var projects []Project
		err := db.Find(&projects).Error
		if err != nil {
			return fmt.Sprintf("Error querying projects: %v", err), false
		}
		var b strings.Builder
		b.WriteString("<span style=\"color:var(--accent-color)\">SYSTEM PROJECTS:</span><br>")
		b.WriteString("<table style=\"width:100%; border-collapse:collapse; margin-top:5px; font-family:'Fira Code', monospace; font-size:0.8rem; border: 1px solid var(--text-primary);\">")
		b.WriteString("<tr style=\"border-bottom: 2px solid var(--text-primary); background: var(--btn-bg);\">")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">ID</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">TITLE</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">TECH</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">CATEGORY</th>")
		b.WriteString("</tr>")
		for _, p := range projects {
			b.WriteString("<tr style=\"border-bottom: 1px solid var(--text-secondary);\">")
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%d</td>", p.ID))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(p.Title)))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(p.Tech)))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(p.Category)))
			b.WriteString("</tr>")
		}
		b.WriteString("</table>")
		return b.String(), true

	case "skills":
		var skills []Skill
		err := db.Find(&skills).Error
		if err != nil {
			return fmt.Sprintf("Error querying skills: %v", err), false
		}
		var b strings.Builder
		b.WriteString("<span style=\"color:var(--accent-color)\">SYSTEM SKILLS:</span><br>")
		b.WriteString("<table style=\"width:100%; border-collapse:collapse; margin-top:5px; font-family:'Fira Code', monospace; font-size:0.8rem; border: 1px solid var(--text-primary);\">")
		b.WriteString("<tr style=\"border-bottom: 2px solid var(--text-primary); background: var(--btn-bg);\">")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">ID</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">NAME</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">LEVEL</th>")
		b.WriteString("</tr>")
		for _, s := range skills {
			b.WriteString("<tr style=\"border-bottom: 1px solid var(--text-secondary);\">")
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%d</td>", s.ID))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(s.Name)))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%d%%</td>", s.Level))
			b.WriteString("</tr>")
		}
		b.WriteString("</table>")
		return b.String(), true

	case "analytics":
		var analytics []Analytics
		err := db.Find(&analytics).Error
		if err != nil {
			return fmt.Sprintf("Error querying analytics: %v", err), false
		}
		var b strings.Builder
		b.WriteString("<span style=\"color:var(--accent-color)\">SITE ANALYTICS DATA:</span><br>")
		b.WriteString("<table style=\"width:100%; border-collapse:collapse; margin-top:5px; font-family:'Fira Code', monospace; font-size:0.8rem; border: 1px solid var(--text-primary);\">")
		b.WriteString("<tr style=\"border-bottom: 2px solid var(--text-primary); background: var(--btn-bg);\">")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">ID</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">EVENT</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">COUNT</th>")
		b.WriteString("</tr>")
		for _, a := range analytics {
			b.WriteString("<tr style=\"border-bottom: 1px solid var(--text-secondary);\">")
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%d</td>", a.ID))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(a.Event)))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%d hits</td>", a.Count))
			b.WriteString("</tr>")
		}
		b.WriteString("</table>")
		return b.String(), true

	case "leaderboard":
		var leaderboard []CTFScore
		err := db.Order("score desc, created_at asc").Find(&leaderboard).Error
		if err != nil {
			return fmt.Sprintf("Error querying leaderboard: %v", err), false
		}
		var b strings.Builder
		b.WriteString("<span style=\"color:var(--accent-color)\">CTF LEADERBOARD SCORES:</span><br>")
		b.WriteString("<table style=\"width:100%; border-collapse:collapse; margin-top:5px; font-family:'Fira Code', monospace; font-size:0.8rem; border: 1px solid var(--text-primary);\">")
		b.WriteString("<tr style=\"border-bottom: 2px solid var(--text-primary); background: var(--btn-bg);\">")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">ID</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">USERNAME</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">SCORE</th>")
		b.WriteString("</tr>")
		for _, l := range leaderboard {
			b.WriteString("<tr style=\"border-bottom: 1px solid var(--text-secondary);\">")
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%d</td>", l.ID))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(l.Username)))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%d PTS</td>", l.Score))
			b.WriteString("</tr>")
		}
		b.WriteString("</table>")
		return b.String(), true

	case "logins":
		var logins []LoginLog
		err := db.Limit(50).Order("created_at desc").Find(&logins).Error
		if err != nil {
			return fmt.Sprintf("Error querying login logs: %v", err), false
		}
		var b strings.Builder
		b.WriteString("<span style=\"color:var(--accent-color)\">RECENT LOGIN HISTORY LOGS (MAX 50):</span><br>")
		b.WriteString("<table style=\"width:100%; border-collapse:collapse; margin-top:5px; font-family:'Fira Code', monospace; font-size:0.8rem; border: 1px solid var(--text-primary);\">")
		b.WriteString("<tr style=\"border-bottom: 2px solid var(--text-primary); background: var(--btn-bg);\">")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">ID</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">USERNAME</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">IP_ADDRESS</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">STATUS</th>")
		b.WriteString("<th style=\"padding:4px 8px; text-align:left; border: 1px solid var(--text-secondary); color: var(--accent-color);\">TIMESTAMP</th>")
		b.WriteString("</tr>")
		for _, l := range logins {
			statusStyle := "color:#22c55e"
			if l.Status != "success" {
				statusStyle = "color:#ef4444"
			}
			b.WriteString("<tr style=\"border-bottom: 1px solid var(--text-secondary);\">")
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%d</td>", l.ID))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(l.Username)))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", htmlEscape(l.IPAddress)))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary); %s\">%s</td>", statusStyle, htmlEscape(l.Status)))
			b.WriteString(fmt.Sprintf("<td style=\"padding:4px 8px; border: 1px solid var(--text-secondary);\">%s</td>", l.CreatedAt.Format("2006-01-02 15:04:05")))
			b.WriteString("</tr>")
		}
		b.WriteString("</table>")
		return b.String(), true

	default:
		return fmt.Sprintf("<span style=\"color:red\">ERROR: Unknown db command: %s. Available subcommands: tables, users, articles, contacts, projects, skills, analytics, leaderboard, logins</span>", htmlEscape(subcmd)), true
	}
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
