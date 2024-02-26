echo "View Sub"
id=$(curl -G https://www.strava.com/api/v3/push_subscriptions \
    -d client_id=$1 \
    -d client_secret=$2 | cut -d ':' -f2 | cut -d ',' -f1)

echo "ID: $id"

echo "Delete Sub"
curl -X DELETE "https://www.strava.com/api/v3/push_subscriptions/$id?client_id=$1&client_secret=$2"

echo "Re-Subscribing"
curl -X POST https://www.strava.com/api/v3/push_subscriptions \
     -F client_id=$1 \
     -F client_secret=$2 \
     -F callback_url=$3/strava/webhook \
     -F verify_token=STRAVA
echo "Done"
