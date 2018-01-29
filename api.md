### API Reference

|Method |Route                                      |Body                               |Allows         |Description                            |
|-------|-------------------------------------------|-----------------------------------|---------------|---------------------------------------|
|GET	|lists/**{userId}**					        |									|owner, friends |Gets all of a user's lists and gifts   |
|POST	|list								        |name, description					|owner			|Creates a list                         |
|POST   |list/**{listId}**					        |name, description					|owner			|Edits a list                           |
|DELETE |list/**{listId}**					        |									|owner			|Removes a list                         |
|       |                                           |                                   |               |                                       |
|POST	|list/**{listId}**/gift				        |name, description, url, imageUrl   |owner			|Creates a gift                         |
|POST	|list/**{listId}**/gift/**{giftId}**		|name, description, url, imageUrl	|owner			|Edits a gift                           |
|DELETE	|list/**{listId}**/gift/**{giftId}**		|									|owner			|Removes a gift                         |
|POST	|list/**{listId}**/gift/**{giftId}**/claim  |state      						|friends		|Claims a gift                          |
|       |                                           |                                   |               |                                       |
|GET	|friends            					    |									|owner          |Gets all of a user's friends           |
|POST	|friend                					    |email								|owner          |Adds a friend                          |
|POST	|friend/accept/**{friendId}**               |   								|friend         |Accepts a friend invite                |
|DELETE	|friend/**{friendId}**             		    |									|owner          |Removes a friend                       |