#!/usr/bin/env bash
################################################################################
# Delete instance of Card Judge in Digital Ocean
################################################################################

set -e # exit on any command error

cd "$(dirname "$0")"

APP_NAME="card-judge"
DROPLET_NAME="$APP_NAME-database"

################################################################################
# delete droplet

echo "----------------------------------------"
echo "Deleting Droplet..."

DROPLET_ID=$(doctl compute droplet list --format=ID,Name --no-header | grep $DROPLET_NAME | cut -d ' ' -f 1)
if [[ -z "$DROPLET_ID" ]]; then
	echo "Droplet ID not found"
else
	read -p "Do you want to backup the database? [Y/n]: " BACKUP_DB
	if [[ "$BACKUP_DB" != "n" ]]; then
		echo "----------------------------------------"
		echo "Backing Up Database..."

		BACKUP_SQL_PATH="$(pwd)/backups/$(date +%Y%m%d%H%M%S)_backup_card_judge.sql"

		DROPLET_IP=$(doctl compute droplet list --format=PublicIPv4,Name --no-header | grep $DROPLET_NAME | cut -d ' ' -f 1)
		if [[ -z "$DROPLET_IP" ]]; then
			echo "Droplet IP not found"
			exit 1
		fi

		ssh root@$DROPLET_IP 'mariadb-dump --order-by-primary CARD_JUDGE | sed -e '\''s/DEFINER[ ]*=[ ]*[^*]*\*/\*/'\'' > /root/backup.sql'
		scp root@$DROPLET_IP:/root/backup.sql "$BACKUP_SQL_PATH" >/dev/null 2>&1

		if [ ! -f "$BACKUP_SQL_PATH" ]; then
			echo "Backup failed: backup file not found"
			exit 1
		fi

		if [ ! -s "$BACKUP_SQL_PATH" ]; then
			echo "Backup failed: backup file is empty"
			exit 1
		fi

		if find "$BACKUP_SQL_PATH" -mmin +1 -print -quit | grep -q .; then
			echo "Backup failed: backup file is older than 1 minute"
			exit 1
		fi

		BACKUP_SQL_SIZE=$(stat -c%s "$BACKUP_SQL_PATH")
		if (( $BACKUP_SQL_SIZE < 1024 )); then
			echo "Backup failed: backup file is too small"
			exit 1
		fi

		BACKUP_SQL_LAST_LINE=$(tail -n 1 "$BACKUP_SQL_PATH")
		if ! [[ "$BACKUP_SQL_LAST_LINE" =~ ^"-- Dump completed on " ]]; then
			echo "Backup failed: backup file does not appear to be valid"
			exit 1
		fi

		BACKUP_GPG_PATH="$BACKUP_SQL_PATH".gpg
		rm -f $BACKUP_GPG_PATH
		gpg -c --output $BACKUP_GPG_PATH $BACKUP_SQL_PATH

		if [ ! -f "$BACKUP_GPG_PATH" ]; then
			echo "File not found: $BACKUP_GPG_PATH"
			exit 1
		fi

		echo "Database Backed Up"
	fi
	doctl compute droplet delete $DROPLET_ID --force
	echo "Droplet Deleted"
fi

################################################################################
# delete app

echo "----------------------------------------"
echo "Deleting App..."

APP_ID=$(doctl apps list --format=ID,Spec.Name --no-header | grep $APP_NAME | cut -d ' ' -f 1)
if [[ -z "$APP_ID" ]]; then
	echo "App ID not found"
else
	doctl apps delete $APP_ID --force
	echo "App Deleted"
fi

################################################################################

exit 0
