# 🔑 OIDC Auth Simulator - Golang

Overview
A lightweight simulator built with Go to demonstrate the OIDC authentication flow using Keycloak as an identity broker and external IdPs like Google or Azure AD.

⚙️ Prerequisites <br>
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
      
  2. Configure Azure AD 
     • In your Azure home dashboard search for app registrations
     • Create a new app registration
     • Click Create Realm with any suitable name and choose supported as accounts in this organisational directory only
     • Scroll down, under Redirect URI choose web and the link has to be pasted from the keycloak. 
     • Now copy the authorization and token endpoints (Oauth 2.0 authorization token endpoint and token endpoint) under endpoints.
     • Copy the client ID and paste it somewhere in notepad that will be used for keycloak configuration.
     • Create a new client secret and copy the client secret and paste it in notepad for configuring the keycloak. 
     
  3. Configure Keycloak
     • Login to Keycloak admin console: http://localhost:8080 → Administration Console
     • Create Realm :
        • Click Create Realm
        • Name: demo (must match OIDC_ISSUER in code)
     • Add an Identity Provider :
        • Choose OpenId Connect v1.0 
        • Enter an alias
        • Copy the redirect URI and use it in 4th step of (2. Configure Azure AD)
        • Choose login flow as First broker login
        • Sync mode as import
        • Use the authorization and token endpoints from the 5th step of(2. Configure Azure AD)
        • Paste the client ID and the client secret from the Azure AD.
        • Mention default scopes as -> openid profile email  
     • Create Client :
        •	Go to: Clients → Create client
	      •	Client ID: go-web-app
	      •	Client type: OpenID Connect
	      •	Root URL: (eg http://localhost:3000)
        •	Save
     • Under Client settings:
	      •	Enable Standard Flow ✅
	      •	Set Valid Redirect URIs to:eg http://localhost:3000/callback
        •	Save
     
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
