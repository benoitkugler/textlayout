package pango

/* Pango Language Sample Table
* Sources:
*
* WP-PANG
* 	Wikipedia's List of Pangrams in Other Languages
* 	http://en.wikipedia.org/wiki/List_of_pangrams#Other_languages
* 	Fetched on 2008-08-19
*
* WP-SFD
* 	Wikipedia's Sample Font Displays in Other Languages
* 	http://en.wikipedia.org/wiki/Sample_Font_Displays_In_Other_Languages
* 	Fetched on 2008-08-19
*
* WP
*      Wikipedia, Article about the language
*      Fetched on 2020-09-08
*
* GLASS
* 	Kermit project's "I Can Eat Glass" list, also available in pango-view/
* 	http://www.columbia.edu/kermit/utf8.html#glass
* 	Fetched on 2008-08-19, updates on 2020-09-08
*
* KERMIT
* 	Kermit project's Quick-Brown-Fox equivalents for other languages
* 	http://www.columbia.edu/kermit/utf8.html#quickbrownfox
* 	Fetched on 2008-08-19
*
* GSPECI
* 	gnome-specimen's translations
* 	http://svn.gnome.org/viewvc/gnome-specimen/trunk/po/
* 	Fetched on 2008-08-19
*
* MISC
* 	Miscellaneous
*
* The sample text may be a pangram, but is not necessarily.  It is chosen to
* be demonstrative of normal text in the language, as well as exposing font
* feature requirements unique to the language.  It should be suitable for use
* as sample text in a font selection dialog.
*
* Needless to say, the list MUST be sorted on the language code.
 */

var langTexts = []languageRecord{
	recordSample{lang: "af", // Afrikaans  GLASS,
		sample: "Ek kan glas eet, maar dit doen my nie skade nie."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "ar", // Arabic  WP - PANG,
		sample: "نص حكيم له سر قاطع وذو شأن عظيم مكتوب على ثوب أخضر ومغلف بجلد أزرق."}, /* A wise text which has an absolute secret and great importance, written on a green tissue and covered with blue leather. */
	recordSample{lang: "arn", // Mapudungun  WP - PANG,
		sample: "Gvxam mincetu apocikvyeh: ñizol ce mamvj ka raq kuse bafkeh mew."}, /* Tale under the full moon: the chief chemamull and the clay old woman at the lake/sea. */
	recordSample{lang: "bar", // Bavarian  GLASS,
		sample: "I koh Glos esa, und es duard ma ned wei."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "bg", // Bulgarian  WP - SFD,
		sample: "Под южно дърво, цъфтящо в синьо, бягаше малко пухкаво зайче."}, /* A little fluffy young rabbit ran under a southern tree blooming in blue */
	recordSample{lang: "bi", // Bislama  GLASS,
		sample: "Mi save kakae glas, hemi no save katem mi."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "bn", // Bengali  GLASS,
		sample: "আমি কাঁচ খেতে পারি, তাতে আমার কোনো ক্ষতি হয় না।"}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "bo", // Tibetan  GLASS,
		sample: "ཤེལ་སྒོ་ཟ་ནས་ང་ན་གི་མ་རེད།"}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "bs", // Bosnian  WP - PANG,
		sample: "Fin džip, gluh jež i čvrst konjić dođoše bez moljca."}, /* A nice jeep, a deaf hedgehog and a tough horse came without a moth. */
	recordSample{lang: "ca", // Catalan  WP - PANG,
		sample: "Jove xef, porti whisky amb quinze glaçons d'hidrogen, coi!"}, /* Young chef, bring whisky with fifteen hydrogen ice cubes, damn! */
	recordSample{lang: "ch", // Chamorro  GLASS,
		sample: "Siña yo' chumocho krestat, ti ha na'lalamen yo'."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "cs", // Czech  WP - SFD,
		sample: "Příliš žluťoučký kůň úpěl ďábelské ódy."}, /* A too yellow horse moaned devil odes. */
	recordSample{lang: "cy", // Welsh  GLASS,
		sample: "Dw i'n gallu bwyta gwydr, 'dyw e ddim yn gwneud dolur i mi."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "da", // Danish  WP - SFD,
		sample: "Quizdeltagerne spiste jordbær med fløde, mens cirkusklovnen Walther spillede på xylofon."}, /* The quiz contestants ate strawberries with cream while Walther the clown was playing the xylophone. */
	recordSample{lang: "de", // German  WP - SFD,
		sample: "Zwölf Boxkämpfer jagen Viktor quer über den großen Sylter Deich."}, /* Twelve boxing fighters drive Viktor over the great. */
	recordSample{lang: "dv", // Maldivian  WP,
		sample: "މާއްދާ 1 – ހުރިހާ އިންސާނުން ވެސް އުފަންވަނީ، ދަރަޖަ އާއި ޙައްޤު ތަކުގައި މިނިވަންކަމާއި ހަމަހަމަކަން ލިބިގެންވާ ބައެއްގެ ގޮތުގައެވެ."}, /* Beginning of UDHR */

	recordSample{lang: "el", // Greek  WP - SFD,
		sample: "Θέλει αρετή και τόλμη η ελευθερία. (Ανδρέας Κάλβος)"}, /* Liberty requires virtue and mettle. (Andreas Kalvos) */
	recordSample{lang: "en", // English  GSPECI,
		sample: "The quick brown fox jumps over the lazy dog."},
	recordSample{lang: "enm", // Middle English  GLASS,
		sample: "Ich canne glas eten and hit hirtiþ me nouȝt."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "eo", // Esperanto  WP - SFD,
		sample: "Eĥoŝanĝo ĉiuĵaŭde."}, /* Change of echo every Thursday. */
	recordSample{lang: "es", // Spanish  WP - PANG,
		sample: "Jovencillo emponzoñado de whisky: ¡qué figurota exhibe!"}, /* Whisky-intoxicated youngster — what a figure he's showing! */
	recordSample{lang: "et", // Estonian  WP - SFD,
		sample: "See väike mölder jõuab rongile hüpata."}, /* This small miller is able to jump on the train. */
	recordSample{lang: "eu", // Basque  GLASS,
		sample: "Kristala jan dezaket, ez dit minik ematen."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "fa", // Persian  MISC, /* Behdad Esfahbod (#548730) */
		sample: "«الا یا اَیُّها السّاقی! اَدِرْ کَأساً وَ ناوِلْها!» که عشق آسان نمود اوّل، ولی افتاد مشکل‌ها!"},
	recordSample{lang: "fi", // Finnish  WP - SFD,
		sample: "Viekas kettu punaturkki laiskan koiran takaa kurkki."}, /* The cunning red-coated fox peeped from behind the lazy dog. */
	recordSample{lang: "fr", // French  MISC, /* Vincent Untz (#549520) http://fr.wikipedia.org/wiki/Pangramme */
		sample: "Voix ambiguë d'un cœur qui, au zéphyr, préfère les jattes de kiwis."}, /* Ambiguous voice of a heart that, in the wind, prefers bowls of kiwis. */
	recordSample{lang: "fro", // Old French  GLASS,
		sample: "Je puis mangier del voirre. Ne me nuit."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "ga", // Irish  WP - PANG,
		sample: "Chuaigh bé mhórshách le dlúthspád fíorfhinn trí hata mo dhea-phorcáin bhig."}, /* A maiden of large appetite with an intensely white, dense spade went through the hat of my good little porker. */
	recordSample{lang: "gd", // Scottish Gaelic  GLASS,
		sample: "S urrainn dhomh gloinne ithe; cha ghoirtich i mi."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "gl", // Galician  GLASS,
		sample: "Eu podo xantar cristais e non cortarme."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "got", // Gothic  GLASS,
		sample: "𐌼𐌰𐌲 𐌲𐌻𐌴𐍃 𐌹̈𐍄𐌰𐌽, 𐌽𐌹 𐌼𐌹𐍃 𐍅𐌿 𐌽𐌳𐌰𐌽 𐌱𐍂𐌹𐌲𐌲𐌹𐌸."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "gu", // Gujarati  GLASS,
		sample: "હું કાચ ખાઇ શકુ છુ અને તેનાથી મને દર્દ નથી થતુ."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "gv", // Manx Gaelic  GLASS,
		sample: "Foddym gee glonney agh cha jean eh gortaghey mee."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "haw", // Hawaiian  GLASS,
		sample: "Hiki iaʻu ke ʻai i ke aniani; ʻaʻole nō lā au e ʻeha."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "he", // Hebrew  WP - SFD,
		sample: "דג סקרן שט לו בים זך אך לפתע פגש חבורה נחמדה שצצה כך."}, /* A curious fish sailed a clear sea, and suddenly found nice company that just popped up. */
	recordSample{lang: "hi", // Hindi  MISC, /* G Karunakar (#549532) */
		sample: "नहीं नजर किसी की बुरी नहीं किसी का मुँह काला जो करे सो उपर वाला"}, /* its not in the sight or the face, but its all in god's grace. */
	recordSample{lang: "hr", // Croatian  MISC,
		sample: "Deblji krojač: zgužvah smeđ filc u tanjušni džepić."}, /* A fatter taylor: I’ve crumpled a brown felt in a slim pocket. */
	recordSample{lang: "hu", // Hungarian  WP - SFD,
		sample: "Egy hűtlen vejét fülöncsípő, dühös mexikói úr Wesselényinél mázol Quitóban."}, /* An angry Mexican man, who caught his faithless son-in-law, is painting Wesselényi's house in Quito. */
	recordSample{lang: "hy", // Armenian  GLASS,
		sample: "Կրնամ ապակի ուտել և ինծի անհանգիստ չըներ։"}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "is", // Icelandic  WP - PANG,
		sample: "Kæmi ný öxi hér ykist þjófum nú bæði víl og ádrepa"}, /* If a new axe were here, thieves would feel increasing deterrence and punishment. */
	recordSample{lang: "it", // Italian  WP - SFD,
		sample: "Ma la volpe, col suo balzo, ha raggiunto il quieto Fido."}, /* But the fox, with its jump, reached the calm dog */
	recordSample{lang: "ja", // Japanese  KERMIT,
		sample: "いろはにほへと ちりぬるを 色は匂へど 散りぬるを"},
	recordSample{lang: "jam", // Jamaican Creole English  KERMIT,
		sample: "Chruu, a kwik di kwik brong fox a jomp huova di liezi daag de, yu no siit?"},
	recordSample{lang: "jbo", // Lojban  WP - PANG,
		sample: ".o'i mu xagji sofybakni cu zvati le purdi"}, /* Watch out, five hungry Soviet-cows are in the garden! */
	recordSample{lang: "jv", // Javanese  GLASS,
		sample: "Aku isa mangan beling tanpa lara."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "ka", // Georgian  GLASS,
		sample: "მინას ვჭამ და არა მტკივა."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "km", // Khmer  GLASS,
		sample: "ខ្ញុំអាចញុំកញ្ចក់បាន ដោយគ្មានបញ្ហារ"}, /* I can eat glass and it doesn't hurt me. */

	recordSample{lang: "kn", // Kannada  GLASS,
		sample: "ನಾನು ಗಾಜನ್ನು ತಿನ್ನಬಲ್ಲೆ ಮತ್ತು ಅದರಿಂದ ನನಗೆ ನೋವಾಗುವುದಿಲ್ಲ."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "ko", // Korean  WP - SFD,
		sample: "다람쥐 헌 쳇바퀴에 타고파"}, /* I Wanna ride on the chipmunk's old hamster wheel. */
	recordSample{lang: "kw", // Cornish  GLASS,
		sample: "Mý a yl dybry gwéder hag éf ny wra ow ankenya."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "la", // Latin  WP - PANG,
		sample: "Sic surgens, dux, zelotypos quam karus haberis"},
	recordSample{lang: "lo", // Lao  GLASS,
		sample: "ຂອ້ຍກິນແກ້ວໄດ້ໂດຍທີ່ມັນບໍ່ໄດ້ເຮັດໃຫ້ຂອ້ຍເຈັບ"}, /* I can eat glass and it doesn't hurt me. */

	recordSample{lang: "lt", // Lithuanian  WP - PANG,
		sample: "Įlinkdama fechtuotojo špaga sublykčiojusi pragręžė apvalų arbūzą."}, /* Incurving fencer sword sparkled and perforated a round watermelon. */
	recordSample{lang: "lv", // Latvian  WP - SFD,
		sample: "Sarkanās jūrascūciņas peld pa jūru."}, /* Red seapigs swim in the sea. */
	recordSample{lang: "map", // Marquesan  GLASS,
		sample: "E koʻana e kai i te karahi, mea ʻā, ʻaʻe hauhau."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "mk", // Macedonian  GLASS,
		sample: "Можам да јадам стакло, а не ме штета."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "ml", // Malayalam  GLASS,
		sample: "വേദനയില്ലാതെ കുപ്പിചില്ലു് എനിയ്ക്കു് കഴിയ്ക്കാം."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "mn", // Mongolian  GLASS,
		sample: "ᠪᠢ ᠰᠢᠯᠢ ᠢᠳᠡᠶᠦ ᠴᠢᠳᠠᠨᠠ ᠂ ᠨᠠᠳᠤᠷ ᠬᠣᠤᠷᠠᠳᠠᠢ ᠪᠢᠰᠢ"}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "mr", // Marathi  GLASS,
		sample: "मी काच खाऊ शकतो, मला ते दुखत नाही."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "ms", // Malay  GLASS,
		sample: "Saya boleh makan kaca dan ia tidak mencederakan saya."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "my", // Burmese  WP,
		sample: "ဘာသာပြန်နှင့် စာပေပြုစုရေး ကော်မရှင်"}, /* Literary and Translation Commission */
	recordSample{lang: "nap", // Neapolitan  GLASS,
		sample: "M' pozz magna' o'vetr, e nun m' fa mal."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "nb", // Norwegian Bokmål  GSPECI,
		sample: "Vår sære Zulu fra badeøya spilte jo whist og quickstep i min taxi."},
	recordSample{lang: "nl", // Dutch  WP - SFD,
		sample: "Pa's wijze lynx bezag vroom het fikse aquaduct."}, /* Dad's wise lynx piously regarded the substantial aqueduct. */
	recordSample{lang: "nn", // Norwegian Nynorsk  GLASS,
		sample: "Eg kan eta glas utan å skada meg."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "no", // Norwegian Bokmål  GSPECI,
		sample: "Vår sære Zulu fra badeøya spilte jo whist og quickstep i min taxi."},
	recordSample{lang: "nv", // Navajo  GLASS,
		sample: "Tsésǫʼ yishą́ągo bííníshghah dóó doo shił neezgai da."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "oc", // Occitan  GLASS,
		sample: "Pòdi manjar de veire, me nafrariá pas."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "or", // Oriya  GLASS,
		sample: "ମୁଁ କାଚ ଖାଇପାରେ ଏବଂ ତାହା ମୋର କ୍ଷତି କରିନଥାଏ।."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "pa", // Punjabi  GLASS,
		sample: "ਮੈਂ ਗਲਾਸ ਖਾ ਸਕਦਾ ਹਾਂ ਅਤੇ ਇਸ ਨਾਲ ਮੈਨੂੰ ਕੋਈ ਤਕਲੀਫ ਨਹੀਂ."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "pcd", // Picard  GLASS,
		sample: "Ch'peux mingi du verre, cha m'foé mie n'ma."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "pl", // Polish  WP - SFD,
		sample: "Pchnąć w tę łódź jeża lub ośm skrzyń fig."}, /* Push into this boat a hedgehog or eight boxes of figs. */
	recordSample{lang: "pt", // Portuguese  WP - SFD,
		sample: "Vejam a bruxa da raposa Salta-Pocinhas e o cão feliz que dorme regalado."}, /* Watch the witch of the Jump-Puddles fox and the happy dog that sleeps delighted. */
	recordSample{lang: "pt-br", // Brazilian Portuguese  WP - PANG,
		sample: "À noite, vovô Kowalsky vê o ímã cair no pé do pingüim queixoso e vovó põe açúcar no chá de tâmaras do jabuti feliz."}, /* At night, grandpa Kowalsky sees the magnet falling in the complaining penguin's foot and grandma puts sugar in the happy tortoise's date tea.*/
	recordSample{lang: "ro", // Romanian  MISC, /* Misu Moldovan (#552993) */
		sample: "Fumegând hipnotic sașiul azvârle mreje în bălți."}, /* Hypnotically smoking, the cross-eyed man throws fishing nets into ponds. */
	recordSample{lang: "ru", // Russian  WP - PANG,
		sample: "В чащах юга жил бы цитрус? Да, но фальшивый экземпляр!"}, /* Would a citrus live in the bushes of the south? Yes, but only a fake one! */
	recordSample{lang: "sa", // Sanskrit  GLASS,
		sample: "काचं शक्नोम्यत्तुम् । नोपहिनस्ति माम् ॥"}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "scn", // Sicilian  GLASS,
		sample: "Puotsu mangiari u vitru, nun mi fa mali."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "si", // Sinhalese  WP,
		sample: "මනොපුබ්‌බඞ්‌ගමා ධම්‌මා, මනොසෙට්‌ඨා මනොමයා; මනසා චෙ පදුට්‌ඨෙන, භාසති වා කරොති වා; තතො නං දුක්‌ඛමන්‌වෙති, චක්‌කංව වහතො පදං."},
	recordSample{lang: "sk", // Slovak  KERMIT,
		sample: "Starý kôň na hŕbe kníh žuje tíško povädnuté ruže, na stĺpe sa ďateľ učí kvákať novú ódu o živote."},
	recordSample{lang: "sl", // Slovenian  WP - PANG,
		sample: "Šerif bo za vajo spet kuhal domače žgance."}, /* For an exercise, sheriff will again make home-made mush. */
	recordSample{lang: "sq", // Albanian  GLASS,
		sample: "Unë mund të ha qelq dhe nuk më gjen gjë."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "sr", // Serbian (Cyrillic)  WP - SFD,
		sample: "Чешће цeђење мрeжастим џаком побољшава фертилизацију генских хибрида."}, /* More frequent filtering through the reticular bag improves fertilization of genetic hybrids. */
	recordSample{lang: "sv", // Swedish  WP - SFD,
		sample: "Flygande bäckasiner söka strax hwila på mjuka tuvor."}, /* Flying snipes soon look to rest on soft grass beds. */
	recordSample{lang: "swg", // Swabian  GLASS,
		sample: "I kå Glas frässa, ond des macht mr nix!"}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "ta", // Tamil  GLASS,
		sample: "நான் கண்ணாடி சாப்பிடுவேன், அதனால் எனக்கு ஒரு கேடும் வராது."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "te", // Telugu  GLASS,
		sample: "నేను గాజు తినగలను అయినా నాకు యేమీ కాదు."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "th", // Thai  WP - SFD,
		sample: "เป็นมนุษย์สุดประเสริฐเลิศคุณค่า - กว่าบรรดาฝูงสัตว์เดรัจฉาน - จงฝ่าฟันพัฒนาวิชาการ อย่าล้างผลาญฤๅเข่นฆ่าบีฑาใคร - ไม่ถือโทษโกรธแช่งซัดฮึดฮัดด่า - หัดอภัยเหมือนกีฬาอัชฌาสัย - ปฏิบัติประพฤติกฎกำหนดใจ - พูดจาให้จ๊ะ ๆ จ๋า ๆ น่าฟังเอยฯ"}, /* Being a man is worthy - Beyond senseless animal - Begin educate thyself - Begone from killing and trouble - Bear not thy grudge, damn and, curse - Bestow forgiving and sporting - Befit with rules - Benign speech speak thou */
	recordSample{lang: "tl", // Tagalog  GLASS,
		sample: "Kaya kong kumain nang bubog at hindi ako masaktan."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "tr", // Turkish  WP - PANG,
		sample: "Pijamalı hasta yağız şoföre çabucak güvendi."}, /* The patient in pajamas trusted the swarthy driver quickly. */
	recordSample{lang: "tw", // Twi  GLASS,
		sample: "Metumi awe tumpan, ɜnyɜ me hwee."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "uk", // Ukrainian  WP - PANG,
		sample: "Чуєш їх, доцю, га? Кумедна ж ти, прощайся без ґольфів!"}, /* Daughter, do you hear them, eh? Oh, you are funny! Say good-bye without knee-length socks. */
	recordSample{lang: "ur", // Urdu  GLASS,
		sample: "میں کانچ کھا سکتا ہوں اور مجھے تکلیف نہیں ہوتی ۔"}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "vec", // Venetian  GLASS,
		sample: "Mi posso magnare el vetro, no'l me fa mae."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "vi", // Vietnamese  GSPECI,
		sample: "Con sói nâu nhảy qua con chó lười."},
	recordSample{lang: "wa", // Walloon  GLASS,
		sample: "Dji pou magnî do vêre, çoula m' freut nén må."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "yi", // Yiddish  GLASS,
		sample: "איך קען עסן גלאָז און עס טוט מיר נישט װײ."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "yo", // Yoruba  GLASS,
		sample: "Mo lè je̩ dígí, kò ní pa mí lára."}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "zh-cn", // Chinese Simplified  GLASS,
		sample: "我能吞下玻璃而不伤身体。"}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "zh-mo", // Chinese Traditional  GLASS,
		sample: "我能吞下玻璃而不傷身體。"}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "zh-sg", // Chinese Simplified  GLASS,
		sample: "我能吞下玻璃而不伤身体。"}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "zh-tw", // Chinese Traditional  GLASS,
		sample: "我能吞下玻璃而不傷身體。"}, /* I can eat glass and it doesn't hurt me. */
	recordSample{lang: "zlm", // Malay  GLASS,
		sample: "Saya boleh makan kaca dan ia tidak mencederakan saya."}, /* I can eat glass and it doesn't hurt me. */
}
