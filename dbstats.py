import statsd, pymongo, time

## This file is intended to be run as a daemon in order to monitor the total number of records available in a variety of
## different mongodb collections. It submits the number of records in each collection to statd as a gauge metric.

def start():
	mongo_client = pymongo.MongoClient("mongo.service.fy", 27017)
	statsd_client = statsd.StatsClient('statsd.service.fy', 8125)

	db = mongo_client.league

	for collection in [db.raw_games, db.raw_leagues, db.raw_summoners, db.processed_games, db.processed_leagues, db.processed_summoners]:
		cnt = collection.count()
		statsd_client.gauge("league.storage." + collection.name, cnt)

if __name__ == "__main__":
	while True:
		start()

		# Take a pulse roughly once per minute.
		time.sleep(60)
