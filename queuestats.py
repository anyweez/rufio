import beanstalkc, statsd, time

## This file is intended to be run as a daemon in order to monitor the total number of records available in a variety of
## different mongodb collections. It submits the number of records in each collection to statd as a gauge metric.

def start():
	statsd_client = statsd.StatsClient('statsd.service.fy', 8125)
	beanstalk_client = beanstalkc.Connection(host='beanstalk.service.fy', port=11300)

	for tube in beanstalk_client.tubes():
		cnt = beanstalk_client.stats_tube(tube)['current-jobs-urgent']
		statsd_client.gauge("league.queue." + tube, cnt)

if __name__ == "__main__":
	while True:
		start()

		# Take a pulse roughly once per minute.
		time.sleep(60)
