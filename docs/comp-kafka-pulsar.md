# Comparatif Kafka vs Pulsar au 19/06/2019

* Architecture :
	* Kafka : 
		* Chaque broker à une partition et dedans il y a le log
		:arrow_right: la façon dont kafka gère les données dans zookeeper limite le scale up
	* Pulsar: ![enter image description here](https://streaml.io/media/img/blog/pulsar-bookkeeper-cluster.png)
		* Chaque broker est assigné à une partition de topic
		* Les messages sont conservés dans les bookies.
		* Dans chaque bookie il y a un distributed log, ce log est décomposé en segments = Apache BookKeeper ledger :arrow_right: les segments sont repartis dans les bookies de sorte à ce que les charges sont bien balancées au sein des bookies présents dans le cluster, ça permet de scale up facilement
	
	![enter image description here](https://streaml.io/media/img/blog/segment-vs-partition.png)


* Spécificité : 
	* Kafka :
	* Pulsar: 
		* segmentation des logs
		* Les urls : `{persistent|non-persistent}://tenant/namespace/topic`
			ex : `"ws://localhost:8080/ws/v2/producer/persistent/public/default/my-topic"` : les topics doivent être organisés en **tenant** et **namespace**. Un namespace contient plusieurs topics et un tenant contient plusieurs namespace
		* 3 modes d'abonnements : 
			*  Exclusive Subscription - Only one consumer can read the topic via the subscription at a time
			*   **Shared Subscription** - Competing consumers can read the topic via the same subscription at the same time.
	    
			*   Fail-Over Subscription - Active/Backup pattern for consumers. If the active consumer dies, then the back up takes over. But there are never two active consumers at the same time.
* Leader :
	* Kafka : Chaque partition a un serveur qui sert de leader, les autres sont des followers qui répliquent le leader. si le leader tombe un autre prend sa place : redondance
	* Pulsar: Chaque bookie a un leader (un broker pulsar) qui lui dit quoi faire 

* Rentention message/TTL : 
	* Kafka : Rentention oui, TLL non
	* Pulsar: Rentention oui, TLL oui
* Offset message : 
	* Kafka : visible facilement via message.offset
chaque consommateur reprend la lecture du début par défaut mais on peut préciser très facilement l'offset de début
	* Pulsar: chaque abonnement possède un curseur (comme un offset pour le consommateur)  
	 Pour les reader on peut spécifier à partir de où on veut lire(msgId ou earliest ou latest) à l'inverse des consommateurs avec lesquels on ne peut pas (on reprend au message le plus récent non acké)
* Création topic :
	* Kafka : besoin de créer les topics à la main (createTopics(...) )
	* Pulsar : pour créer un topic on précise l'url comme vu précédemment 
* Partitionage topic :
	* Kafka : choix du nb de partitions pour un topic à la création du topic
	* Pulsar: choix du nb de partitions possible mais par défaut un topic à un broker (ça limite le flux)
* Configuration possible :
	* Kafka : choix du replicationfactor
	* Pulsar:

* Autre : 
	* Kafka :
	* Pulsar:
# Sources
## Kafka
Documentation officielle kafka
[Doc](http://kafka.apache.org/documentation.html)
[Article scientifique](https://www.microsoft.com/en-us/research/wp-content/uploads/2017/09/Kafka.pdf)

[Résumé](https://www.freecodecamp.org/news/what-to-consider-for-painless-apache-kafka-integration-df559e828876/)

## Pulsar
Documentation officielle pulsar
[Doc](https://pulsar.apache.org/docs/en/standalone/)
[Source](https://github.com/apache/pulsar)



[Intro à Pulsar](https://blog.univalence.io/un-apercu-dapache-pulsar/) 12/06/19
[Résumé Pulsar](https://jack-vanlightly.com/blog/2018/10/2/understanding-how-apache-pulsar-works) 13/06/19



## Comparatif

[apache-kafka-vs-apache-pulsar](https://blog.univalence.io/apache-kafka-vs-apache-pulsar/)
12/06/19

https://stackoverflow.com/questions/46048608/what-are-the-advantages-and-disadvantages-of-kafka-over-apache-pulsar 12/06/19
****************************

[apache-kafka-vs-apache-pulsar](https://streaml.io/blog/pulsar-streaming-queuing) 12/06/19

https://streaml.io/blog/pulsar-segment-based-architecture  12/06/19


