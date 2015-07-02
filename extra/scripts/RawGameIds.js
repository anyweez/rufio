/**
 * This script generates a list of all partially-known game ID's that are present in raw_games
 * logs. All elements of the output list are unique.
 */
conn = new Mongo();
db = conn.getDB("league")

cursor = db.raw_games.find({

}, {
	"response.games.gameid": true,
});

// Store game ID's in a set so that they're deduped in constant time.
gameids = {};
	
while (cursor.hasNext()) {
	record = cursor.next();

	for (var i = 0; i < record.response.games.length; i++) {
		gameids[record.response.games[i].gameid] = true
	}
}

for (k in gameids) {
	print(k)
}
//printjson(gameids)
