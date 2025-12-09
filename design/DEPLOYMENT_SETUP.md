# Dalleserver Deployment Setup

This guide covers the one-time setup required to deploy dalleserver to a Digital Ocean instance using systemd and GitHub Actions CI/CD.

## Target Environment

- **Server IP:** 167.71.187.196
- **User:** jrush
- **Repo Location:** ~/Development/trueblocks-dalleserver
- **Service:** dalleserver
- **Port:** 8081
- **Process Manager:** systemd

## Prerequisites

- SSH access to 167.71.187.196 as user `jrush`
- Git installed on the remote machine
- Go 1.25.1+ installed on the remote machine
- Basic familiarity with systemd and SSH

## Phase 1: Manual Setup on 167.71.187.196

### Step 1: Clone the Repository

```bash
cd ~/Development
git clone https://github.com/TrueBlocks/trueblocks-dalleserver.git
cd trueblocks-dalleserver
```

### Step 2: Build and Test Locally

```bash
make build
```

Verify the binary was created:

```bash
ls -la trueblocks-dalleserver
```

### Step 3: Create Environment File

Create an environment file in your home directory to store secrets (keep this private, never commit to git):

```bash
# Create the file with restricted permissions
touch ~/.dalleserver.env
chmod 600 ~/.dalleserver.env
```

Edit the file and add your configuration variables (e.g., `OPENAI_API_KEY`, etc.).

### Step 4: Create Systemd Service File

Create `/etc/systemd/system/dalleserver.service`:

```bash
sudo tee /etc/systemd/system/dalleserver.service > /dev/null <<'EOF'
[Unit]
Description=TrueBlocks DALLE Server
After=network.target

[Service]
Type=simple
User=jrush
WorkingDirectory=/home/jrush/Development/trueblocks-dalleserver
EnvironmentFile=%h/.dalleserver.env
ExecStart=/home/jrush/Development/trueblocks-dalleserver/trueblocks-dalleserver --port=8081
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
```

This configures:
- Service runs as user `jrush`
- Automatically starts on boot
- Listens on port 8081 (to avoid conflict with IPFS on 8080)
- Restarts if it crashes (with 10-second delay)
- Logs to systemd journal (viewable via `journalctl`)
- Loads environment variables from a private config file in your home directory (`%h` expands to home directory)

### Step 5: Enable and Start the Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable dalleserver
sudo systemctl start dalleserver
```

### Step 6: Verify the Service is Running

```bash
sudo systemctl status dalleserver
```

You should see:
- `Active: active (running)`
- Process ID listed

### Step 7: Check Logs

```bash
journalctl -u dalleserver -f
```

This shows live logs. Press `Ctrl+C` to exit.

### Step 8: Test the Server

From another terminal:

```bash
curl http://167.71.187.196:8081/health
```

You should get a response from the server.

## Managing the Service

### View Status

```bash
sudo systemctl status dalleserver
```

### View Recent Logs

```bash
journalctl -u dalleserver -n 50
```

### View Live Logs

```bash
journalctl -u dalleserver -f
```

### Stop the Service

```bash
sudo systemctl stop dalleserver
```

### Start the Service

```bash
sudo systemctl start dalleserver
```

### Restart the Service

```bash
sudo systemctl restart dalleserver
```

## Phase 2: GitHub Actions Setup (Automated Deployment)

Once this manual setup is complete on 167.71.187.196, the GitHub Actions workflow will:

1. Detect when you create a git tag (e.g., `git tag v1.2.3`)
2. SSH into 167.71.187.196
3. Pull the tagged code
4. Rebuild the binary
5. Restart the systemd service

See `.github/workflows/deploy-dalleserver.yml` for the full workflow configuration.

### Initial GitHub Actions Setup

A deploy SSH key will be generated and added to GitHub Secrets. This allows the CI/CD to SSH into the server without manual intervention.

## Troubleshooting

### Service won't start

Check the logs:

```bash
journalctl -u dalleserver -n 100
```

Common issues:
- Binary path incorrect in service file
- Working directory doesn't exist
- Environment variables not set correctly
- Port already in use

### Permission denied errors

Verify file ownership:

```bash
ls -la /home/jrush/trueblocks-minidapps/dalleserver/
```

If needed, fix permissions:

```bash
chmod +x /home/jrush/trueblocks-minidapps/dalleserver/trueblocks-dalleserver
```

### Port 8080 already in use

Check what's using the port:

```bash
sudo lsof -i :8080
```

Either stop the other service or change dalleserver's port in the service file.

### Environment variables not loading

Verify the `.env` file exists and is readable:

```bash
cat ~/.dalleserver.env
```

Check that `EnvironmentFile=/home/jrush/.dalleserver.env` is in the service file:

```bash
sudo cat /etc/systemd/system/dalleserver.service
```

After editing the service file, reload and restart:

```bash
sudo systemctl daemon-reload
sudo systemctl restart dalleserver
```

## Phase 2: GitHub Actions CI/CD Setup

A GitHub Actions workflow has been configured to automatically deploy new versions when you tag the repository.

### Workflow Trigger

The workflow triggers on git tags matching the pattern `v*` (e.g., `v1.2.3`, `v2.0.0`).

### How to Deploy

To deploy a new version:

```bash
# Make and commit your changes
git add .
git commit -m "Your message"

# Create a tag
git tag v1.2.3

# Push the tag to GitHub
git push origin v1.2.3
```

GitHub Actions will automatically:
1. Check out the tagged version
2. Build the binary with version information from the tag
3. SSH into 167.71.187.196
4. Pull the tagged code
5. Rebuild the binary
6. Restart the systemd service

### Monitoring Deployments

Watch the deployment progress in GitHub:

1. Go to your repository on GitHub
2. Click the **Actions** tab
3. Select the running "Deploy to Production" workflow
4. View real-time logs as the deployment progresses

### GitHub Secrets Required

The workflow uses `secrets.DEPLOY_SSH_KEY` to authenticate to the server. This must be configured in your GitHub repository settings:

1. Go to **Settings** → **Secrets and variables** → **Actions**
2. Create a secret named `DEPLOY_SSH_KEY`
3. Paste the contents of the `id_rsa` private key

### Sudo Without Password

The workflow runs `sudo systemctl` commands without a password. This is configured on the server with:

```bash
sudo visudo
```

And adding:

```
jrush ALL=(ALL) NOPASSWD: /bin/systemctl
```

## Future Updates

Once this setup is complete, you can deploy new versions by simply tagging the repo:

```bash
git tag v1.2.3
git push origin v1.2.3
```

GitHub Actions will automatically build, deploy, and restart the service on 167.71.187.196.

You can monitor the deployment in the GitHub Actions tab of your repository.
