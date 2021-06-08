import freetype

fonts = [
    "test/4x6.pcf",
    "test/8x16.pcf.gz",
    "test/charB18.pcf.gz",
    "test/courB18.pcf.gz",
    "test/hanglg16.pcf.gz",
    "test/helvB18.pcf.gz",
    "test/lubB18.pcf.gz",
    "test/ncenB18.pcf.gz",
    "test/orp-italic.pcf.gz",
    "test/timB18.pcf.gz",
    "test/timR24-ISO8859-1.pcf.gz",
    "test/timR24.pcf.gz",
]

print("var expectedSizes = []Size{")
for file in fonts:
    face = freetype.Face(file)
    size = face.available_sizes[0]
    print(f"    {{Height: {size.height}, Width: {size.width}, XPpem: {int(size.x_ppem / 64)}, YPpem: {int(size.y_ppem / 64)} }},")
print("}")
