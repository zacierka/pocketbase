import requests

url = 'http://localhost:8090/strava/webhook'
myobj = {
    "object_type": "activity",
    "object_id": ,
    "aspect_type": "create",
    "owner_id": ,
    "event_time": 1516126040,
    "subscription_id": 
}

x = requests.post(url, json = myobj)

print(x.text)