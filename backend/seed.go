package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	mrand "math/rand"
	"regexp"
	"strings"
)

func randomRef() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

const keyAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func newKey() string {
	parts := make([]string, 3)
	for i := range parts {
		var sb strings.Builder
		for j := 0; j < 5; j++ {
			n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(keyAlphabet))))
			sb.WriteByte(keyAlphabet[n.Int64()])
		}
		parts[i] = sb.String()
	}
	return strings.Join(parts, "-")
}

type game struct {
	name, genre string
	year        int
	priceEUR    int // typical list price in whole euros
	desc        string
}

var games = []game{
	{"Cyberpunk 2077", "RPG", 2020, 60, "Open-world action RPG set in Night City, a megalopolis obsessed with power, glamour and body modification."},
	{"The Witcher 3: Wild Hunt", "RPG", 2015, 40, "Hunt monsters as Geralt of Rivia in a vast, award-winning open world."},
	{"Elden Ring", "RPG", 2022, 60, "Explore the Lands Between in FromSoftware's open-world action RPG, co-written with George R. R. Martin."},
	{"Baldur's Gate 3", "RPG", 2023, 60, "Gather your party and return to the Forgotten Realms in a tale of fellowship and betrayal."},
	{"Red Dead Redemption 2", "Action", 2018, 60, "An epic tale of life in America's unforgiving heartland."},
	{"Grand Theft Auto V", "Action", 2013, 30, "Three criminals, one city, and the heist of a lifetime."},
	{"FC 25", "Sports", 2024, 70, "The world's game, rebuilt with HyperMotionV and Rush mode."},
	{"NBA 2K25", "Sports", 2024, 70, "Build your legend on the court with ProPLAY technology."},
	{"Call of Duty: Black Ops 6", "Shooter", 2024, 70, "Spy-thriller campaign, omnimovement multiplayer and Zombies."},
	{"Counter-Strike 2", "Shooter", 2023, 15, "The legendary tactical shooter rebuilt on Source 2."},
	{"Hogwarts Legacy", "Adventure", 2023, 60, "Live the Wizarding World in an immersive open-world action RPG."},
	{"Starfield", "RPG", 2023, 70, "Bethesda's first new universe in 25 years — explore it with the crew of Constellation."},
	{"Hades II", "Roguelike", 2024, 30, "Battle beyond the Underworld as Melinoë, Princess of the Underworld."},
	{"Hollow Knight: Silksong", "Platformer", 2025, 20, "Hunt your way through a kingdom ruled by silk and song."},
	{"Stardew Valley", "Simulation", 2016, 15, "Inherit a farm and build the life of your dreams in Pelican Town."},
	{"Minecraft", "Sandbox", 2011, 30, "Build, explore and survive in infinite procedurally generated worlds."},
	{"Terraria", "Sandbox", 2011, 10, "Dig, fight, explore and build in this 2D adventure."},
	{"Resident Evil 4", "Horror", 2023, 60, "Survival horror reimagined — Leon's mission to rescue the president's daughter."},
	{"Resident Evil Village", "Horror", 2021, 40, "Experience survival horror like never before in the eighth major instalment."},
	{"Dark Souls III", "RPG", 2016, 60, "Embrace the darkness in the final chapter of the Dark Souls trilogy."},
	{"Sekiro: Shadows Die Twice", "Action", 2019, 60, "Carve your own clever path to vengeance in Sengoku-era Japan."},
	{"God of War Ragnarök", "Action", 2022, 70, "Kratos and Atreus journey through the Nine Realms as Ragnarök approaches."},
	{"Marvel's Spider-Man 2", "Action", 2023, 70, "Swing as both Peter Parker and Miles Morales across Marvel's New York."},
	{"The Last of Us Part I", "Adventure", 2022, 70, "Experience the emotional journey of Joel and Ellie, rebuilt for modern hardware."},
	{"Horizon Forbidden West", "Action", 2022, 60, "Join Aloy as she braves the Forbidden West."},
	{"Ghost of Tsushima", "Action", 2020, 60, "Forge a new path and wage an unconventional war for the freedom of Tsushima."},
	{"Forza Horizon 5", "Racing", 2021, 60, "Your ultimate Horizon adventure awaits in vibrant Mexico."},
	{"Gran Turismo 7", "Racing", 2022, 70, "The real driving simulator returns."},
	{"Need for Speed Unbound", "Racing", 2022, 70, "Race, outrun the cops and rise through the street racing scene in Lakeshore."},
	{"EA Sports WRC", "Racing", 2023, 50, "Drive the world's most thrilling rally cars on iconic stages."},
	{"Halo Infinite", "Shooter", 2021, 60, "The Master Chief returns in the most expansive Halo campaign yet."},
	{"Gears 5", "Shooter", 2019, 40, "Kait Diaz fights to uncover her connection to the enemy."},
	{"Doom Eternal", "Shooter", 2020, 40, "Rip and tear through hell's armies in id Software's blistering shooter."},
	{"Borderlands 3", "Shooter", 2019, 30, "Loot, shoot and cause chaos in the ultimate looter-shooter."},
	{"Far Cry 6", "Shooter", 2021, 40, "Become a guerrilla fighter and liberate the island of Yara."},
	{"Assassin's Creed Mirage", "Action", 2023, 50, "Return to the roots of the series in 9th-century Baghdad."},
	{"Assassin's Creed Valhalla", "Action", 2020, 50, "Lead Viking raids and settle in the lands of England."},
	{"Diablo IV", "RPG", 2023, 70, "The endless battle between Heaven and Hell rages on."},
	{"Overwatch 2", "Shooter", 2022, 15, "Team-based hero shooter with 5v5 combat."},
	{"World of Warcraft: The War Within", "MMO", 2024, 50, "Journey beneath Azeroth in the 10th expansion."},
	{"Hearthstone Bundle", "Card", 2014, 10, "Collect cards and build the ultimate deck."},
	{"Civilization VI", "Strategy", 2016, 60, "Build an empire to stand the test of time."},
	{"Age of Empires IV", "Strategy", 2021, 40, "One of the most beloved RTS games returns."},
	{"Total War: Warhammer III", "Strategy", 2022, 60, "The ultimate Total War fantasy sandbox."},
	{"XCOM 2", "Strategy", 2016, 50, "Lead the resistance against alien occupation."},
	{"Cities: Skylines II", "Simulation", 2023, 50, "Raise a city from the ground up."},
	{"The Sims 4", "Simulation", 2014, 40, "Unleash your imagination and create a unique world."},
	{"Euro Truck Simulator 2", "Simulation", 2012, 20, "Travel across Europe as king of the road."},
	{"Microsoft Flight Simulator", "Simulation", 2020, 70, "Fly the whole planet in stunning detail."},
	{"It Takes Two", "Co-op", 2021, 40, "A genre-bending platform adventure made purely for co-op."},
	{"Lethal Company", "Co-op", 2023, 10, "Scavenge scrap from abandoned moons to meet the quota."},
	{"Helldivers 2", "Shooter", 2024, 40, "Spread managed democracy across the galaxy."},
	{"Palworld", "Survival", 2024, 30, "Fight, farm, build and work alongside mysterious creatures."},
	{"Valheim", "Survival", 2021, 20, "A brutal exploration and survival game for 1–10 Vikings."},
	{"Subnautica", "Survival", 2018, 30, "Descend into the depths of an alien underwater world."},
	{"Rust", "Survival", 2018, 40, "The only aim in Rust is to survive."},
	{"Phasmophobia", "Horror", 2020, 15, "A 4-player online co-op psychological horror."},
	{"Dead by Daylight", "Horror", 2016, 20, "A multiplayer horror game where one player is the killer."},
	{"Silent Hill 2", "Horror", 2024, 70, "Fear and tragedy return in the remake of the horror classic."},
	{"Alan Wake 2", "Horror", 2023, 60, "A string of ritualistic murders threatens Bright Falls."},
	{"Control Ultimate Edition", "Action", 2019, 40, "A supernatural third-person action-adventure."},
	{"Death Stranding Director's Cut", "Adventure", 2019, 40, "Reconnect a fractured world, one delivery at a time."},
	{"Disco Elysium", "RPG", 2019, 40, "A groundbreaking role-playing game with unprecedented freedom."},
	{"Persona 5 Royal", "RPG", 2022, 60, "Don the mask and join the Phantom Thieves of Hearts."},
	{"Final Fantasy VII Rebirth", "RPG", 2024, 70, "The next chapter in the reimagined FINAL FANTASY VII."},
	{"Final Fantasy XVI", "RPG", 2023, 70, "An epic dark fantasy of Eikons and Dominants."},
	{"Dragon's Dogma 2", "RPG", 2024, 70, "Become the Arisen and embark on an adventure."},
	{"Monster Hunter: World", "Action", 2018, 30, "Hunt colossal monsters in an ever-changing ecosystem."},
	{"Monster Hunter Wilds", "Action", 2025, 70, "Hunt in a world of dynamic, changing climates."},
	{"Street Fighter 6", "Fighting", 2023, 60, "The next evolution of the legendary fighting series."},
	{"Mortal Kombat 1", "Fighting", 2023, 70, "Experience a reborn Mortal Kombat Universe."},
	{"Tekken 8", "Fighting", 2024, 70, "The next chapter of the Mishima family saga."},
	{"Super Mario Odyssey", "Platformer", 2017, 60, "Join Mario on a massive 3D adventure around the world."},
	{"The Legend of Zelda: Tears of the Kingdom", "Adventure", 2023, 70, "An epic adventure across the land and skies of Hyrule."},
	{"Mario Kart 8 Deluxe", "Racing", 2017, 60, "Race friends or battle them in a revised battle mode."},
	{"Animal Crossing: New Horizons", "Simulation", 2020, 60, "Escape to a deserted island and create your own paradise."},
	{"Pokémon Scarlet", "RPG", 2022, 60, "Discover a new region of Paldea."},
	{"Metroid Dread", "Action", 2021, 50, "Samus returns in the final chapter of the Metroid saga."},
	{"Hades", "Roguelike", 2020, 25, "Defy the god of the dead as you hack and slash out of the Underworld."},
	{"Dead Cells", "Roguelike", 2018, 25, "A roguelite, metroidvania-inspired action platformer."},
	{"Slay the Spire", "Roguelike", 2019, 25, "Craft a unique deck, encounter bizarre creatures and slay the Spire."},
	{"Celeste", "Platformer", 2018, 20, "Help Madeline survive her inner demons on her climb up Celeste Mountain."},
	{"Cuphead", "Platformer", 2017, 20, "A classic run-and-gun action game in 1930s cartoon style."},
	{"Hollow Knight", "Platformer", 2017, 15, "Explore twisting caverns, battle tainted creatures and befriend bizarre bugs."},
	{"Ori and the Will of the Wisps", "Platformer", 2020, 30, "Embark on an all-new adventure in a vast, exotic world."},
	{"Portal 2", "Puzzle", 2011, 10, "The 'Perpetual Testing Initiative' has been expanded."},
	{"Half-Life: Alyx", "Adventure", 2020, 60, "Valve's VR return to the Half-Life universe."},
	{"Left 4 Dead 2", "Shooter", 2009, 10, "Cooperative zombie survival on the road to the rescue."},
	{"Team Fortress 2 Mann Co. Pack", "Shooter", 2007, 10, "Nine distinct classes provide a broad range of tactical abilities."},
	{"Apex Legends Coins", "Shooter", 2019, 10, "Conquer with character in this free-to-play hero shooter."},
	{"Fortnite V-Bucks", "Shooter", 2017, 10, "Spend V-Bucks on items in Battle Royale."},
	{"Rainbow Six Siege", "Shooter", 2015, 20, "Tactical team-based shooter of precise breaching."},
	{"The Division 2", "Shooter", 2019, 30, "Defend Washington D.C. in this online open-world shooter RPG."},
	{"Destiny 2: Lightfall", "Shooter", 2023, 50, "Wield the power of Strand against the Witness."},
	{"Warhammer 40,000: Space Marine 2", "Action", 2024, 60, "Embody the superhuman power of a Space Marine."},
	{"Black Myth: Wukong", "Action", 2024, 60, "An action RPG rooted in Chinese mythology."},
	{"Lies of P", "Action", 2023, 60, "A Soulslike twist on the tale of Pinocchio."},
	{"Hi-Fi Rush", "Action", 2023, 30, "A rhythm action game where the world moves to the beat."},
	{"Sea of Thieves", "Adventure", 2018, 40, "The Pirate Adventure Game — sail, explore and plunder."},
	{"No Man's Sky", "Adventure", 2016, 60, "Explore an infinite procedurally generated universe."},
	{"Satisfactory", "Simulation", 2024, 40, "Build massive factories on an alien planet."},
	{"Factorio", "Simulation", 2020, 35, "Build and maintain factories."},
	{"RimWorld", "Simulation", 2018, 35, "A sci-fi colony sim driven by an intelligent AI storyteller."},
	{"Crusader Kings III", "Strategy", 2020, 50, "Dynastic grand strategy in the Middle Ages."},
	{"Stellaris", "Strategy", 2016, 40, "Explore a galaxy full of wonders in this sci-fi grand strategy."},
	{"Anno 1800", "Strategy", 2019, 50, "Lead the Industrial Revolution."},
	{"Football Manager 2024", "Sports", 2023, 50, "Take charge of your club's future."},
	{"F1 24", "Racing", 2024, 70, "Drive in the pinnacle of motorsport."},
	{"eFootball Coins", "Sports", 2022, 15, "Free-to-play football simulation."},
	{"Rocket League", "Sports", 2015, 20, "Soccer meets driving in this high-powered hybrid."},
	{"Football Manager Touch", "Sports", 2023, 20, "Manage on the go."},
	{"Tomb Raider: Definitive Survivor Trilogy", "Adventure", 2024, 40, "Experience Lara Croft's origin trilogy."},
	{"Mass Effect Legendary Edition", "RPG", 2021, 60, "Relive the Mass Effect trilogy remastered."},
	{"Dragon Age: The Veilguard", "RPG", 2024, 70, "Lead a team of companions to stop two ancient evils."},
	{"Fallout 4", "RPG", 2015, 30, "Bethesda's post-nuclear open-world masterpiece."},
	{"The Elder Scrolls V: Skyrim Anniversary", "RPG", 2011, 40, "Endless adventure awaits in the 10th-anniversary edition."},
	{"Oblivion Remastered", "RPG", 2025, 50, "Return to Cyrodiil, reimagined in Unreal Engine 5."},
	{"Kingdom Come: Deliverance II", "RPG", 2025, 60, "Return to medieval Bohemia in this realistic RPG."},
	{"Clair Obscur: Expedition 33", "RPG", 2025, 50, "A turn-based RPG with real-time mechanics in Belle Époque France."},
	{"Split Fiction", "Co-op", 2025, 50, "Two writers get trapped in their own stories."},
	{"Schedule I", "Simulation", 2025, 20, "Build your drug empire from the ground up."},
	{"Peak", "Co-op", 2025, 8, "Climb a mountain with friends — and try not to fall."},
}

var platforms = []struct {
	name   string
	factor float64
}{
	{"Steam", 1.0}, {"Xbox", 1.05}, {"PlayStation", 1.08}, {"Nintendo", 1.1}, {"EA App", 1.0}, {"Epic Games", 0.98}, {"Battle.net", 1.0}, {"GOG", 0.95},
}

var regions = []string{"Global", "Europe", "North America", "United Kingdom"}

var sellerNames = []string{"KeyForge", "GameVault", "PixelDepot", "LootLlama", "DigitalDirect", "NordKeys", "BlueFox Games", "CDKeyHub", "ArcadeAlley", "TurboKeys"}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	return strings.Trim(nonAlnum.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

// platformFor limits a game to the stores where it plausibly exists.
func platformsFor(g game, rng *mrand.Rand) []string {
	n := g.name
	switch {
	case strings.Contains(n, "Zelda") || strings.Contains(n, "Mario") || strings.Contains(n, "Animal Crossing") || strings.Contains(n, "Pokémon") || strings.Contains(n, "Metroid"):
		return []string{"Nintendo"}
	case strings.Contains(n, "God of War") || strings.Contains(n, "Spider-Man") || strings.Contains(n, "Last of Us") || strings.Contains(n, "Horizon") || strings.Contains(n, "Ghost of") || strings.Contains(n, "Gran Turismo"):
		return []string{"PlayStation"}
	case strings.Contains(n, "Halo") || strings.Contains(n, "Gears") || strings.Contains(n, "Forza") || strings.Contains(n, "Sea of Thieves") || strings.Contains(n, "Ori and"):
		return []string{"Xbox", "Steam"}
	case strings.Contains(n, "Diablo") || strings.Contains(n, "Overwatch") || strings.Contains(n, "Warcraft") || strings.Contains(n, "Hearthstone") || strings.Contains(n, "Black Ops"):
		return []string{"Battle.net", "Xbox", "PlayStation"}
	case strings.Contains(n, "FC 25") || strings.Contains(n, "Need for Speed") || strings.Contains(n, "Sims") || strings.Contains(n, "Dragon Age") || strings.Contains(n, "Mass Effect") || strings.Contains(n, "Apex") || strings.Contains(n, "WRC") || strings.Contains(n, "F1 24"):
		return []string{"EA App", "Steam", "Xbox", "PlayStation"}
	case strings.Contains(n, "Witcher") || strings.Contains(n, "Cyberpunk") || strings.Contains(n, "Disco Elysium") || strings.Contains(n, "Baldur"):
		return []string{"Steam", "GOG", "Xbox", "PlayStation"}
	}
	all := []string{"Steam", "Xbox", "PlayStation", "Epic Games", "GOG"}
	rng.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
	k := 1 + rng.Intn(3)
	out := all[:k]
	hasSteam := false
	for _, p := range out {
		if p == "Steam" {
			hasSteam = true
		}
	}
	if !hasSteam && rng.Intn(3) > 0 {
		out = append(out, "Steam")
	}
	return out
}

func (a *App) seed(ctx context.Context, force bool) error {
	var n int
	if err := a.db.QueryRow(ctx, `SELECT count(*) FROM products`).Scan(&n); err != nil {
		return err
	}
	if n > 0 && !force {
		return nil
	}
	rng := mrand.New(mrand.NewSource(42))
	tx, err := a.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if force {
		if _, err := tx.Exec(ctx, `TRUNCATE keys, order_items, orders, offers, sellers, products, banners RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
	}

	sellerIDs := []int64{}
	for _, s := range sellerNames {
		var id int64
		rating := 4.0 + rng.Float64()*0.99
		if err := tx.QueryRow(ctx, `INSERT INTO sellers(name,rating,sales_count) VALUES($1,$2,$3) RETURNING id`, s, fmt.Sprintf("%.2f", rating), 500+rng.Intn(60000)).Scan(&id); err != nil {
			return err
		}
		sellerIDs = append(sellerIDs, id)
	}

	pf := map[string]float64{}
	for _, p := range platforms {
		pf[p.name] = p.factor
	}

	for gi, g := range games {
		// older games sell more units in aggregate; the first entries get a popularity boost
		popularity := 20000 - gi*120 + rng.Intn(4000)
		for _, plat := range platformsFor(g, rng) {
			region := regions[rng.Intn(len(regions))]
			if rng.Intn(10) < 6 {
				region = "Global"
			}
			list := int(float64(g.priceEUR*100) * pf[plat])
			if list < 300 {
				list = 300
			}
			list = list/100*100 + 99 // 59.99 style prices
			slug := slugify(g.name + " " + plat + " " + region)
			var pid int64
			if err := tx.QueryRow(ctx, `INSERT INTO products(slug,name,platform,region,genre,description,list_price_cents,release_year,sales)
				VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
				slug, g.name, plat, region, g.genre, g.desc, list, g.year, popularity/(1+rng.Intn(3))+rng.Intn(500)).Scan(&pid); err != nil {
				return err
			}
			// 2–6 sellers compete, always under the list price (that is the marketplace's selling point)
			perm := rng.Perm(len(sellerIDs))
			offerCount := 2 + rng.Intn(5)
			for _, idx := range perm[:offerCount] {
				discount := 0.08 + rng.Float64()*0.55
				price := int(float64(list) * (1 - discount))
				price = price/10*10 + 9
				if price < 99 {
					price = 99
				}
				stock := 3 + rng.Intn(10)
				var oid int64
				if err := tx.QueryRow(ctx, `INSERT INTO offers(product_id,seller_id,price_cents,stock) VALUES($1,$2,$3,$4) RETURNING id`, pid, sellerIDs[idx], price, stock).Scan(&oid); err != nil {
					return err
				}
				for k := 0; k < stock; k++ {
					if _, err := tx.Exec(ctx, `INSERT INTO keys(offer_id,code) VALUES($1,$2)`, oid, newKey()); err != nil {
						return err
					}
				}
			}
		}
	}

	banners := []struct{ title, sub, cta, link, a, b string }{
		{"Biggest deals on game keys", "Compare offers from trusted sellers and save up to 70%", "Browse deals", "/store?sort=discount&onSale=1", "#4c1d95", "#7c3aed"},
		{"New releases, lower prices", "Pre-load the newest games at marketplace prices", "See new releases", "/store?sort=newest", "#0f766e", "#14b8a6"},
		{"Play on Xbox & PlayStation", "Console keys delivered instantly after checkout", "Shop console", "/store?platform=Xbox,PlayStation", "#9a3412", "#f97316"},
	}
	for i, b := range banners {
		if _, err := tx.Exec(ctx, `INSERT INTO banners(title,subtitle,cta,link,color_a,color_b,position) VALUES($1,$2,$3,$4,$5,$6,$7)`, b.title, b.sub, b.cta, b.link, b.a, b.b, i); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
