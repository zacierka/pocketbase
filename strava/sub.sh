curl -X POST https://www.strava.com/api/v3/push_subscriptions \
      -F client_id=$1 \
      -F client_secret=$2 \
      -F callback_url=$3/strava/webhook \
      -F verify_token=STRAVA