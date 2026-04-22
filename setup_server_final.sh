#!/bin/bash

export DEBIAN_FRONTEND=noninteractive

sudo apt-get update -y
sudo apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release podman podman-compose

sudo rm -f /usr/share/keyrings/doppler-archive-keyring.gpg
curl -sLf --retry 5 --retry-delay 2 --connect-timeout 10 --tlsv1.2 --proto "=https" 'https://packages.doppler.com/public/cli/gpg.DE2A7741A397C129.key' | sudo gpg --dearmor -o /usr/share/keyrings/doppler-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/doppler-archive-keyring.gpg] https://packages.doppler.com/public/cli/deb/debian any-version main" | sudo tee /etc/apt/sources.list.d/doppler-cli.list
sudo apt-get update -y
sudo apt-get install -y doppler

if ! command -v doppler &> /dev/null; then
    curl -sLf --retry 5 https://cli.doppler.com/install.sh | sudo sh
fi

mkdir -p /home/gatsu51/perjadin_data/{db,redis,uploads}
mkdir -p /home/gatsu51/perjadin-app
sudo chown -R gatsu51:gatsu51 /home/gatsu51/perjadin_data
sudo chown -R gatsu51:gatsu51 /home/gatsu51/perjadin-app
sudo chmod -R 755 /home/gatsu51/perjadin_data

sudo loginctl enable-linger gatsu51

sudo -u gatsu51 bash << 'EOF'
cd /home/gatsu51/perjadin-app
doppler configure set token dp.st.prd.7aPgQ5OuqBr3HdwX4PzoPQ91t8PkQ0TcFOVzXysGCKw --scope /home/gatsu51/perjadin-app
echo "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" | podman login ghcr.io -u xm4yestiK --password-stdin

mkdir -p /home/gatsu51/.ssh
chmod 700 /home/gatsu51/.ssh
if [ ! -f /home/gatsu51/.ssh/id_ed25519 ]; then
    ssh-keygen -t ed25519 -f /home/gatsu51/.ssh/id_ed25519 -N ""
fi
cat /home/gatsu51/.ssh/id_ed25519.pub >> /home/gatsu51/.ssh/authorized_keys
chmod 600 /home/gatsu51/.ssh/authorized_keys

git config --global --add safe.directory /home/gatsu51/perjadin-app

echo "=== PRIVATE KEY FOR GITHUB SECRETS (PROD_SSH_KEY) ==="
cat /home/gatsu51/.ssh/id_ed25519
echo "===================================================="
EOF
