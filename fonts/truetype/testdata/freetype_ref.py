import freetype

fonts = [
    "testdata/ToyFeat.ttf",
    "testdata/ToySbix.ttf",
    "testdata/ToyCBLC1.ttf",
    "testdata/ToyCBLC2.ttf",
    "testdata/NotoColorEmoji.ttf",
    "testdata/mry_KacstQurn.ttf",
    "testdata/IBM3161-bitmap.otb",
    "testdata/Gacha_9.dfont",
]

print("var expectedSizes = [][]Size{")
for file in fonts:
    face = freetype.Face(file)
    print("{")
    for size in face.available_sizes:
        print(f"    {{Height: {size.height}, Width: {size.width}, XPpem: {int(size.x_ppem / 64)}, YPpem: {int(size.y_ppem / 64)} }},")
    print("},")
print("}")
