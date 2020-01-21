from datetime import datetime, timedelta

time = datetime.utcnow().replace(microsecond=0) +timedelta(minutes=1)
print(time.isoformat()+"Z")
