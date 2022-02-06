package engine

import (
	"errors"
	"instantchat.rooms/instantchat/backend/internal/domain_structures"
	"instantchat.rooms/instantchat/backend/internal/util"
	"github.com/google/uuid"
	"math/rand"
	"net/url"
	"strings"
)

const RoomUserNameMinLength = 1
const RoomUserNameMaxLength = 80

var ProvidedNameTaken = errors.New("provided name already taken")
var BadNameLength = errors.New("provided name must be between " + string(RoomUserNameMinLength) + " and " + string(RoomUserNameMaxLength) + " characters")

var AnonNames = []string{
	"Aardvark",
	"Abyssinian cat",
	"Adelie Penguin",
	"Affenpinscher",
	"Airedale Terrier",
	"Airedoodle",
	"Akbash",
	"Akita",
	"Alabai",
	"Alaskan Husky",
	"Alaskan Malamute",
	"Alaskan Shepherd",
	"Albacore Tuna",
	"Albatross",
	"Aldabra Giant Tortoise",
	"Alligator",
	"Alpaca",
	"Alpine Dachsbracke",
	"Alpine Goat",
	"Alusky",
	"Amur Leopard",
	"Anchovies",
	"Angelfish",
	"Angora Goat",
	"Ant",
	"Anteater",
	"Antelope",
	"Appenzeller Dog",
	"Apple Head Chihuahua",
	"Arapaima",
	"Arctic Fox",
	"Arctic Hare",
	"Arctic Wolf",
	"Armadillo",
	"Asian Elephant",
	"Hornet",
	"Asian Palm Civet",
	"Asiatic Black Bear",
	"Aurochs",
	"Aussiedoodle",
	"Aussiedor",
	"Avocet",
	"Axolotl",
	"Aye Aye",
	"Babirusa",
	"Bactrian Camel",
	"Badger",
	"Baiji",
	"Balinese",
	"Banded Palm Civet",
	"Bandicoot",
	"Banjo Catfish",
	"Barb",
	"Barn Owl",
	"Barnacle",
	"Barracuda",
	"Barramundi Fish",
	"Barred Owl",
	"Basenji Dog",
	"Basking Shark",
	"Bassador",
	"Basset Hound",
	"Bassetoodle",
	"Bat",
	"Bea-Tzu",
	"Beabull",
	"Beagador",
	"Beagle",
	"Beaglier",
	"Beago",
	"Bear",
	"Beaski",
	"Beaver",
	"Bedlington Terrier",
	"Beefalo",
	"Beetle",
	"Bergamasco",
	"Bernedoodle",
	"Bichon Frise",
	"Biewer Terrier",
	"Bighorn Sheep",
	"Bilby",
	"Binturong",
	"Bird",
	"Bird Of Paradise",
	"Birman cat",
	"Bison",
	"Blister Beetle",
	"Blobfish",
	"Bloodhound",
	"Blue grosbeak",
	"Blue Iguana",
	"Blue Jay",
	"Blue Lacy Dog",
	"Blue Shark",
	"Blue Whale",
	"Bluefin Tuna",
	"Bluetick Coonhound",
	"Boar",
	"Bobcat",
	"Boggle",
	"Boglen Terrier",
	"Bolognese Dog",
	"Bongo",
	"Bonito Fish",
	"Bonnethead Shark",
	"Booby",
	"Borador",
	"Border Collie",
	"Border Terrier",
	"Bordoodle",
	"Borkie",
	"Borneo Elephant",
	"Boston Terrier",
	"Bottlenose Dolphin",
	"Bowfin",
	"Bowhead Whale",
	"Box Turtle",
	"Boxador",
	"Boxer Dog",
	"Boxerdoodle",
	"Boxsky",
	"Boxweiler",
	"Boykin Spaniel",
	"Brown Bear",
	"Brown Hyena",
	"Budgerigar",
	"Buffalo",
	"Bull Shark",
	"Bull Terrier",
	"Bulldog",
	"Bullfrog",
	"Bullmastiff",
	"Bumblebee",
	"Burmese",
	"Burmese Python",
	"Burrowing Frog",
	"Bush Baby",
	"Butterfly",
	"Butterfly Fish",
	"Caiman",
	"Caiman Lizard",
	"Cairn Terrier",
	"Camel",
	"Camel Cricket",
	"Camel Spider",
	"Canaan Dog",
	"Canada Lynx",
	"Canadian Eskimo Dog",
	"Canadian Horse",
	"Cane Corso",
	"Capybara",
	"Caracal",
	"Carolina Dog",
	"Carolina Parakeet",
	"Carp",
	"Cashmere Goat",
	"Cassowary",
	"Cat",
	"Caterpillar",
	"Catfish",
	"Cavador",
	"Cavapoo",
	"Cesky Fousek",
	"Cesky Terrier",
	"Chameleon",
	"Chamois",
	"Cheagle",
	"Cheetah",
	"Chickadee",
	"Chicken",
	"Chihuahua",
	"Chimaera",
	"Chinchilla",
	"Chinstrap Penguin",
	"Chipmunk",
	"Chipoo",
	"Chow Chow",
	"Chow Shepherd",
	"Cicada",
	"Cichlid",
	"Clouded Leopard",
	"Clownfish",
	"Clumber Spaniel",
	"Coati",
	"Cockapoo",
	"Cockatoo",
	"Codfish",
	"Coelacanth",
	"Collared Peccary",
	"Collie",
	"Colossal Squid",
	"Buzzard",
	"Loon",
	"Raven",
	"Toad",
	"Cookiecutter Shark",
	"Corgidor",
	"Corgipoo",
	"Corkie",
	"Corman Shepherd",
	"Cotton-top Tamarin",
	"Cougar",
	"Cow",
	"Coyote",
	"Crab",
	"Crab Spider",
	"Crane",
	"Crested Penguin",
	"Crocodile",
	"Curly Coated Retriever",
	"Cuscus",
	"Cuttlefish",
	"Dachsador",
	"Dachshund",
	"Dalmadoodle",
	"Dalmador",
	"Dalmatian",
	"Dapple Dachshund",
	"Deer",
	"Deer Head Chihuahua",
	"Desert Locust",
	"Desert Rain Frog",
	"Desert Tortoise",
	"Deutsche Bracke",
	"Dhole",
	"Dingo",
	"Discus",
	"Doberman Pinscher",
	"Dodo",
	"Dog",
	"Dogo Argentino",
	"Dogue De Bordeaux",
	"Dolphin",
	"Dormouse",
	"Double Doodle",
	"Doxiepoo",
	"Doxle",
	"Dragonfish",
	"Dragonfly",
	"Drever",
	"Drum Fish",
	"Duck",
	"Dugong",
	"Dunker",
	"Dusky Dolphin",
	"Dwarf Crocodile",
	"Eagle",
	"Eastern Bluebird",
	"Eastern Phoebe",
	"Echidna",
	"Edible Frog",
	"Eel",
	"Egyptian Mau",
	"Electric Eel",
	"Elephant",
	"Elephant Seal",
	"Elephant Shrew",
	"Elk",
	"Emperor Penguin",
	"Emperor Tamarin",
	"Emu",
	"English Cocker Spaniel",
	"English Cream Golden Retriever",
	"English Pointer",
	"English Shepherd",
	"English Springer Spaniel",
	"Entlebucher Mountain Dog",
	"Epagneul Pont Audemer",
	"Ermine",
	"Eskimo Dog",
	"Eskipoo",
	"Estrela Mountain Dog",
	"Falcon",
	"Fallow deer",
	"False Killer Whale",
	"Fangtooth",
	"Feist",
	"Fennec Fox",
	"Ferret",
	"Ferruginous Hawk",
	"Field Spaniel",
	"Fin Whale",
	"Finnish Spitz",
	"Fire salamander",
	"Fire-Bellied Toad",
	"Fish",
	"Fisher Cat",
	"Flamingo",
	"Flat-Coated Retriever",
	"Florida Gar",
	"Florida Panther",
	"Flounder",
	"Flying Fish",
	"Flying Lemur",
	"Flying Squirrel",
	"Fossa",
	"Fox",
	"Fox Terrier",
	"French Bulldog",
	"Frengle",
	"Frilled Lizard",
	"Frilled Shark",
	"Frog",
	"Fruit Bat",
	"Fur Seal",
	"Galapagos Penguin",
	"Galapagos Tortoise",
	"Gar",
	"Gecko",
	"Gentoo Penguin",
	"Geoffroys Tamarin",
	"Gerberian Shepsky",
	"Gerbil",
	"German Pinscher",
	"German Shepherd Guide",
	"German Sheppit",
	"German Sheprador",
	"German Spitz",
	"Gharial",
	"Giant African Land Snail",
	"Giant Armadillo",
	"Giant Clam",
	"Giant Panda Bear",
	"Giant Salamander",
	"Giant Schnauzer",
	"Giant Schnoodle",
	"Gila Monster",
	"Giraffe",
	"Glass Frog",
	"Glass Lizard",
	"Glechon",
	"Glen Of Imaal Terrier",
	"Goat",
	"Goberian",
	"Goblin Shark",
	"Goldador",
	"Golden Dox",
	"Golden Lion Tamarin",
	"Golden Masked Owl",
	"Golden Newfie",
	"Golden Oriole",
	"Golden Pyrenees",
	"Golden Retriever",
	"Golden-Crowned Flying Fox",
	"Goldendoodle",
	"Goliath Frog",
	"Goose",
	"Gopher",
	"Gouldian Finch",
	"Grasshopper",
	"Grasshopper Mouse",
	"Gray Fox",
	"Gray Tree Frog",
	"Great Dane",
	"Great Danoodle",
	"Great Pyrenees",
	"Great White Shark",
	"Greater Swiss Mountain Dog",
	"Green Bee-Eater",
	"Green Frog",
	"Green Tree Frog",
	"Greenland Dog",
	"Greenland Shark",
	"Grey Reef Shark",
	"Grey Seal",
	"Greyhound",
	"Griffonshire",
	"Grizzly Bear",
	"Grouse",
	"Guinea Fowl",
	"Guppy",
	"Hagfish",
	"Hammerhead Shark",
	"Hamster",
	"Harbor Seal",
	"Hare",
	"Harp Seal",
	"Harpy Eagle",
	"Harrier",
	"Havanese",
	"Havapoo",
	"Havashire",
	"Hawaiian Crow",
	"Hedgehog",
	"Hercules Beetle",
	"Hermit Crab",
	"Heron",
	"Herring",
	"Highland Cattle",
	"Hippopotamus",
	"Hoary Bat",
	"Honduran White Bat",
	"Honey Badger",
	"Honey Bee",
	"Hoopoe",
	"Horgi",
	"Horn Shark",
	"Hornbill",
	"Horned Frog",
	"Horned Lizard",
	"Horse",
	"Horseshoe Crab",
	"House Finch",
	"Humboldt Penguin",
	"Hummingbird",
	"Humpback Whale",
	"Huntsman Spider",
	"Huskador",
	"Husky Jack",
	"Huskydoodle",
	"Hyena",
	"Ibex",
	"Ibis",
	"Ibizan Hound",
	"Iguana",
	"Immortal Jellyfish",
	"Impala",
	"Imperial Moth",
	"Indian Elephant",
	"Indian Giant Squirrel",
	"Indian Palm Squirrel",
	"Indian Rhinoceros",
	"Indian Star Tortoise",
	"Indochinese Tiger",
	"Indri",
	"Irish Doodle",
	"Irish Setter",
	"Irish Terrier",
	"Irish WolfHound",
	"Italian Greyhound",
	"Ivory-billed woodpecker",
	"Jackabee",
	"Jackal",
	"Jaguar",
	"Japanese Chin",
	"Japanese Squirrel",
	"Japanese Terrier",
	"Javan Rhinoceros",
	"Jellyfish",
	"Jerboa",
	"Kakapo",
	"Kangaroo",
	"Kangaroo Rat",
	"Keel-Billed Toucan",
	"Keeshond",
	"Kerry Blue Terrier",
	"Kiko Goat",
	"Killer Whale",
	"Kinder Goat",
	"King Cobra",
	"King Crab",
	"King Penguin",
	"Kingfisher",
	"Kinkajou",
	"Kiwi",
	"Koala",
	"Kodkod",
	"Komodo Dragon",
	"Kooikerhondje",
	"Kookaburra",
	"Koolie",
	"Krill",
	"Kudu",
	"Kuvasz",
	"Labmaraner",
	"Labradane",
	"Labradoodle",
	"Labrador Retriever",
	"Labraheeler",
	"Ladybug",
	"Lakeland Terrier",
	"LaMancha Goat",
	"Leaf-Tailed Gecko",
	"Lemming",
	"Leopard",
	"Leopard Cat",
	"Leopard Frog",
	"Leopard Seal",
	"Leopard Tortoise",
	"Lhasapoo",
	"Liger",
	"Lion",
	"Lionfish",
	"Little Brown Bat",
	"Little Penguin",
	"Lizard",
	"Llama",
	"Loach",
	"Lobster",
	"Locust",
	"Long-Eared Owl",
	"Longnose Gar",
	"Lorikeet",
	"Lungfish",
	"Lynx",
	"Macaroni Penguin",
	"Macaw",
	"Magellanic Penguin",
	"Magpie",
	"Maine Coon",
	"Malayan Civet",
	"Malayan Tiger",
	"Mallard",
	"Malteagle",
	"Maltese",
	"Maltipoo",
	"Manatee",
	"Manchester Terrier",
	"Maned Wolf",
	"Manta Ray",
	"Marabou Stork",
	"Marble Fox",
	"Marine Iguana",
	"Marine Toad",
	"Markhor",
	"Marmot",
	"Marsh Frog",
	"Masked Palm Civet",
	"Mastador",
	"Mastiff",
	"Meagle",
	"Meerkat",
	"Megalodon",
	"Mexican Alligator Lizard",
	"Mexican Free-Tailed Bat",
	"Milkfish",
	"Mini Labradoodle",
	"Miniature Bull Terrier",
	"Miniature Husky",
	"Mink",
	"Minke Whale",
	"Mole",
	"Molly",
	"Monarch Butterfly",
	"Mongoose",
	"Mongrel",
	"Monitor Lizard",
	"Monkfish",
	"Monte Iberia Eleuth",
	"Moorhen",
	"Moose",
	"Moray Eel",
	"Morkie",
	"Moth",
	"Mountain Bluebird",
	"Mountain Cur",
	"Mountain Feist",
	"Mountain Lion",
	"Mourning Dove",
	"Mouse",
	"Mule",
	"Muskox",
	"Muskrat",
	"Myna Bird",
	"Narwhal",
	"Neapolitan Mastiff",
	"Newt",
	"Nightingale",
	"Nile Crocodile",
	"Norfolk Terrier",
	"North American Black Bear",
	"Northern Cardinal",
	"Northern Inuit Dog",
	"Norwich Terrier",
	"Nubian Goat",
	"Numbat",
	"Nurse Shark",
	"Ocelot",
	"Octopus",
	"Okapi",
	"Olm",
	"Opossum",
	"Ostrich",
	"Otter",
	"Otterhound",
	"Oyster",
	"Paddlefish",
	"Pademelon",
	"Painted Turtle",
	"Pangolin",
	"Panther",
	"Parrot",
	"Parson Russell Terrier",
	"Patterdale Terrier",
	"Peacock",
	"Peagle",
	"Peekapoo",
	"Pekingese",
	"Pelican",
	"Penguin",
	"Pere Davids Deer",
	"Peregrine Falcon",
	"Petite Goldendoodle",
	"Pheasant",
	"Pied Tamarin",
	"Pigeon",
	"Pika",
	"Pike Fish",
	"Pileated Woodpecker",
	"Pink Fairy Armadillo",
	"Piranha",
	"Pitador",
	"Pitsky",
	"Platypus",
	"Pocket Beagle",
	"Poison Dart Frog",
	"Polar Bear",
	"Polish Lowland Sheepdog",
	"Pomapoo",
	"Pomeagle",
	"Pomsky",
	"Pond Skater",
	"Poochon",
	"Poodle",
	"Pool Frog",
	"Porcupine",
	"Porpoise",
	"Possum",
	"Prairie Dog",
	"Prairie Rattlesnake",
	"Prawn",
	"Pronghorn",
	"Pudelpointer",
	"Pufferfish",
	"Puffin",
	"Pug",
	"Pugapoo",
	"Puggle",
	"Pugshire",
	"Puma",
	"Purple Emperor Butterfly",
	"Purple Finch",
	"Puss Moth",
	"Pygmy Hippopotamus",
	"Pygmy Marmoset",
	"Pygora Goat",
	"Pyrador",
	"Pyredoodle",
	"Quagga",
	"Quail",
	"Quetzal",
	"Quokka",
	"Quoll",
	"Rabbit",
	"Raccoon",
	"Raccoon Dog",
	"Radiated Tortoise",
	"Ragdoll cat",
	"Rat Terrier",
	"Rattlesnake",
	"Red Finch",
	"Red Fox",
	"Red Knee Tarantula",
	"Red Panda",
	"Red Squirrel",
	"Red Wolf",
	"Red-handed Tamarin",
	"Red-winged blackbird",
	"Reindeer",
	"Rhinoceros",
	"River Turtle",
	"Rock Hyrax",
	"Rockfish",
	"Rockhopper Penguin",
	"Rodents",
	"Rose-breasted Grosbeak",
	"Roseate Spoonbill",
	"Rottsky",
	"Rottweiler",
	"Royal Penguin",
	"Ruby-Throated Hummingbird",
	"Russian Bear Dog",
	"Russian Blue",
	"Saarloos Wolfdog",
	"Saber-Toothed Tiger",
	"Sable",
	"Saiga",
	"Saint Berdoodle",
	"Saint Bernard",
	"Saint Shepherd",
	"Salamander",
	"Salmon",
	"Salmon Shark",
	"Samoyed",
	"Sand Lizard",
	"Sand Tiger Shark",
	"Saola",
	"Sardines",
	"Sawfish",
	"Scarlet Macaw",
	"Schneagle",
	"Schnoodle",
	"Scimitar-horned Oryx",
	"Scorpion",
	"Scorpion Fish",
	"Scottish Terrier",
	"Sea Dragon",
	"Sea Lion",
	"Sea Otter",
	"Sea Slug",
	"Sea Squirt",
	"Sea Turtle",
	"Sea Urchin",
	"Seahorse",
	"Seal",
	"Sealyham Terrier",
	"Senegal Parrot",
	"Serval",
	"Shark",
	"Sharp-Tailed Snake",
	"Sheep",
	"Sheepadoodle",
	"Shepadoodle",
	"Shepkita",
	"Shepweiler",
	"Shiba Inu",
	"Shih Poo",
	"Shih Tzu",
	"Shoebill Stork",
	"Shollie",
	"Shrimp",
	"Siamese cat",
	"Siamese Fighting Fish",
	"Siberian Husky",
	"Siberian Retriever",
	"Siberian Tiger",
	"Siberpoo",
	"Silky Terrier",
	"Silver Labrador",
	"Sixgill shark",
	"Skate Fish",
	"Skink Lizard",
	"Skipjack Tuna",
	"Skunk",
	"Skye Terrier",
	"Sleeper Shark",
	"Sloth",
	"Smooth Fox Terrier",
	"Snail",
	"Snake",
	"Snapping Turtle",
	"Snorkie",
	"Snow Leopard",
	"Snowshoe",
	"Snowshoe Hare",
	"Snowy Owl",
	"South China Tiger",
	"Spadefoot Toad",
	"Spanador",
	"Spanish Goat",
	"Sparrow",
	"Spectacled Bear",
	"Sperm Whale",
	"Spinner Shark",
	"Spiny Dogfish",
	"Spixs Macaw",
	"Sponge",
	"Spotted Gar",
	"Springador",
	"Springbok",
	"Springerdoodle",
	"Squid",
	"Squirrel",
	"Sri Lankan Elephant",
	"Staffordshire Bull Terrier",
	"Stag Beetle",
	"Starfish",
	"Stingray",
	"Stoat",
	"Striped Hyena",
	"Striped Rocket Frog",
	"Sturgeon",
	"Sucker Fish",
	"Sugar Glider",
	"Sulcata Tortoise",
	"Sumatran Elephant",
	"Sumatran Rhinoceros",
	"Sumatran Tiger",
	"Sun Bear",
	"Swai Fish",
	"Swan",
	"Swedish Vallhund",
	"Syrian Hamster",
	"Taco Terrier",
	"Tamaskan",
	"Tang",
	"Tapir",
	"Tarpon",
	"Tarsier",
	"Tasmanian Devil",
	"Tasmanian Tiger",
	"Tawny Owl",
	"Termite",
	"Terrier",
	"Tetra",
	"Texas Heeler",
	"Thai Ridgeback",
	"Thorny Devil",
	"Tibetan Fox",
	"Tibetan Mastiff",
	"Tibetan Spaniel",
	"Tibetan Terrier",
	"Tiger",
	"Tiger Salamander",
	"Tiger Shark",
	"Toadfish",
	"Torkie",
	"Tortoise",
	"Toucan",
	"Toy Fox Terrier",
	"Toy Poodle",
	"Tree Frog",
	"Tree Kangaroo",
	"Tree swallow",
	"Tropicbird",
	"Tuatara",
	"Tuna",
	"Turkey",
	"Turkish Angora",
	"Uguisu",
	"Umbrellabird",
	"Utonagan",
	"Vampire Bat",
	"Vampire Squid",
	"Vaquita",
	"Vulture",
	"Wallaby",
	"Walleye Fish",
	"Walrus",
	"Wandering Albatross",
	"Warthog",
	"Wasp",
	"Water Buffalo",
	"Water Dragon",
	"Water Vole",
	"Weasel",
	"Weimardoodle",
	"Welsh Corgi",
	"Welsh Terrier",
	"Westiepoo",
	"Whale Shark",
	"Whippet",
	"Whoodle",
	"Whooping Crane",
	"Wild Boar",
	"Wildebeest",
	"Wire Fox Terrier",
	"Wolf",
	"Wolf Eel",
	"Wolf Spider",
	"Wolffish",
	"Wolverine",
	"Wombat",
	"Wood Bison",
	"Wood Frog",
	"Wood Turtle",
	"Woodpecker",
	"Woolly Mammoth",
	"Wrasse",
	"Wyoming Toad",
	"X-Ray Tetra",
	"Xerus",
	"Yak",
	"Yakutian Laika",
	"Yellow-Eyed Penguin",
	"Yellowfin Tuna",
	"Yoranian",
	"Yorkie Bichon",
	"Yorkie Poo",
	"Yorkshire Terrier",
	"Zebra",
	"Zebra Finch",
	"Zebra Shark",
	"Zebu",
	"Zonkey",
	"Zorse",
}

//must be executed under room lock
func validateOrPickRoomUserName(providedUserName string, room *domain_structures.Room) (string, bool, error) {

	//validate and check name provided by user (if any)
	if providedUserName != "" {
		userNameDecoded, _ := url.QueryUnescape(providedUserName)

		if len([]rune(userNameDecoded)) < RoomUserNameMinLength || len([]rune(userNameDecoded)) > RoomUserNameMaxLength {
			return "", false, BadNameLength
		}

		for _, authorizedUser := range room.AllRoomAuthorizedUsersBySessionUUID {
			if strings.ToLower(authorizedUser.UserName) == strings.ToLower(providedUserName) {
				return "", false, ProvidedNameTaken
			}
		}

		return providedUserName, false, nil
	}

	//pick random anonymous name
	takenUserNames := make(map[string]bool)

	for _, user := range room.AllRoomAuthorizedUsersBySessionUUID {
		if user.IsAnonName {
			takenUserNames[user.UserName] = true
		}
	}

	var filteredAnonNames []string

	for i := 0; i < len(AnonNames); i++ {
		name := AnonNames[i]

		if _, found := takenUserNames[name]; !found {
			filteredAnonNames = append(filteredAnonNames, name)
		}
	}

	if filteredAnonNames == nil {
		userNameUUID, err := uuid.NewUUID()

		if err != nil {
			util.LogSevere("failed to generate UUID: '%s'", err)

			return "", false, err
		}

		return "user-" + userNameUUID.String(), true, nil
	}

	return filteredAnonNames[rand.Intn(len(filteredAnonNames))], true, nil
}
