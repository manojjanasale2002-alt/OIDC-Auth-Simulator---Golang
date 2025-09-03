🔑 Keycloak as OIDC Broker ☁️ Using Google 🌐 & Azure 🔷 IdPs for SSO 🔒

⚙️ Prerequisites
	•	Go 1.21+ installed
	•	Docker installed (for Keycloak)
	•	Docker Compose installed
	•	Browser to test login

🚀 Setup & Run
  1. Start Keycloak (Docker)
     • Spin up Keycloak using Docker Compose: docker compose up -d
     • Runs Keycloak on http://localhost:8080
     • Default credentials for Keycloak :
  	    •	Username: admin
  	    •	Password: admin
     
  2. Configure Keycloak
     • Login to Keycloak admin console: http://localhost:8080 → Administration Console
     • Create Realm :
        • Click Create Realm
        • Name: demo-realm (must match OIDC_ISSUER in code)
     • Create Client :
        •	Go to: Clients → Create client
	      •	Client ID: go-web-app
	      •	Client type: OpenID Connect
	      •	Root URL: (eg http://localhost:3000)
        •	Save
     • Under Client settings:
	      •	Enable Standard Flow ✅
	      •	Set Valid Redirect URIs to:
          eg http://localhost:3000/callback
        •	Save
     • (Optional) Create Test User
	      •	Go to Users → Add User
	      •	Username: demo-user
	      •	Set email, first/last name if you want
	      •	Go to Credentials tab → Set a password (turn OFF temporary)

      ✅ Now Keycloak is ready

  3. Run Go Web App
     • Start the Go web app: go run main.go
     • App runs on (eg http://localhost:3000)

🔐 Login Flow
  1.	Visit http://localhost:3000 → you’ll see a Login with Keycloak button.
  2.	Click → you’ll be redirected to Keycloak login.
  3.	Enter the Keycloak user credentials (e.g., demo-user).
  4.	Keycloak redirects back to /callback on your Go app.
  5.	The app:
    	•	Exchanges the authorization code for tokens (Access, ID, Refresh).
    	•	Verifies the ID token signature.
    	•	Displays:
    	•	Access token (truncated)
    	•	Full ID token
    	•	Decoded claims (JSON map)

🛠️ Configuration
   The Go app reads config from environment variables (with defaults):
   You can override them like:
     >> export OIDC_ISSUER=http://localhost:8080/realms/myrealm
     >> export OIDC_CLIENT_ID=my-app
     >> go run main.go

📖 How It Works (Quick Primer)
	1.	User clicks login → app generates:
    	•	state (CSRF protection)
    	•	code_challenge (PKCE)
	2.	Redirects user to Keycloak authorization endpoint.
	3.	After login, Keycloak redirects back with an authorization code.
	4.	The Go app:
    	•	Exchanges code + code_verifier for tokens.
    	•	Verifies ID token signature (issuer, audience).
    	•	Extracts and shows user claims.

  This is the OIDC Authorization Code Flow with PKCE.

🧹 Cleanup
  To stop Keycloak: docker compose down

✅ With these steps, one should be able to pull the repo, run Keycloak + Go app, and test OIDC login end-to-end.

