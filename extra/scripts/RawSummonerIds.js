/**
 * This script generates a list of all known summoner ID's that are present in raw_games
 * logs. All elements of the output list are unique.
 */
conn = new Mongo();
db = conn.getDB("league")

cursor = db.raw_games.find({

}, {
	"response.games.fellowplayers.summonerid": true,
	"response.summonerid": true
});

// Store game ID's in a set so that they're deduped in constant time.
sids = {};
	
while (cursor.hasNext()) {
	record = cursor.next();

	sids[record.response.summonerid] = true;
	for (var i = 0; i < record.response.games.length; i++) {
		for (var j = 0; j < record.response.games[i].fellowplayers.length; j++) {
			sids[record.response.games[i].fellowplayers[j].summonerid] = true			
		}
	}
}

for (k in sids) {
	print(k)
}
