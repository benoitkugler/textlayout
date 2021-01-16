package fontconfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGetFonts(t *testing.T) {
	fs := FcConfigGetCurrent().fonts[FcSetSystem]
	for _, p := range fs {
		file, res := p.FcPatternObjectGetString(FC_FILE, 0)
		if res != FcResultMatch {
			t.Error("file not present")
		}
		fmt.Println(file)
	}
}

func copyFile(dir, source string) error {
	input, err := ioutil.ReadFile(source)
	if err != nil {
		return err
	}
	dest := filepath.Join(dir, filepath.Base(source))
	err = ioutil.WriteFile(dest, input, 0644)
	return err
}

const FONTFILE = "test/4x6.pcf"

// TODO: implement when cache & scan are OK
func TestCacheDir(t *testing.T) {
	// FcChar8 *fontdir = NULL, *cachedir = NULL;
	// char *basedir,
	// char cmd[512];
	tconf := `<fontconfig>
	  <dir>%s</dir>
	  <cachedir>%s</cachedir>
	</fontconfig>`
	// int ret = 0;
	// FcFontSet *fs;
	// FcPattern *pat;

	basedir, err := ioutil.TempDir("", "bz106632-*")
	if err != nil {
		t.Fatal(err)
	}

	fontdir := filepath.Join(basedir, "fonts")
	cachedir := filepath.Join(basedir, "cache")
	if err := os.Mkdir(fontdir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(cachedir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(fontdir, FONTFILE); err != nil {
		t.Fatal(err)
	}

	config := NewFcConfig()
	conf := fmt.Sprintf(tconf, fontdir, cachedir)

	err = config.ParseAndLoadFromMemory([]byte(conf))
	if err != nil {
		t.Fatal(err)
	}

	//     if (!FcConfigBuildFonts (config))    {
	// 	printf ("E: unable to build fonts\n");
	// 	ret = 1;
	// 	goto bail;
	//     }
	//     fprintf (stderr, "D: Obtaining fonts information\n");
	//     pat = FcPatternCreate ();
	//     fs = FcFontList (config, pat, NULL);
	//     FcPatternDestroy (pat);
	//     if (!fs || fs.nfont != 1)   {
	// 	printf ("E: Unexpected the number of fonts: %d\n", !fs ? -1 : fs.nfont);
	// 	ret = 1;
	// 	goto bail;
	//     }
	//     FcFontSetDestroy (fs);
	//     fprintf (stderr, "D: Removing %s\n", fontdir);
	//     snprintf (cmd, 512, "sleep 1; rm -f %s%s*; sleep 1", fontdir, FC_DIR_SEPARATOR_S);
	//      system (cmd);
	//     fprintf (stderr, "D: Reinitializing\n");
	//     if (FcConfigUptoDate(config))  {
	// 	fprintf (stderr, "E: Config reports up-to-date\n");
	// 	ret = 2;
	// 	goto bail;
	//     }
	//     if (!FcInitReinitialize ())   {
	// 	fprintf (stderr, "E: Unable to reinitialize\n");
	// 	ret = 3;
	// 	goto bail;
	//     }
	//     if (FcConfigGetCurrent () == config)    {
	// 	fprintf (stderr, "E: config wasn't reloaded\n");
	// 	ret = 3;
	// 	goto bail;
	//     }
	//     FcConfigDestroy (config);

	//     config = FcConfigCreate ();
	//     if (!FcConfigParseAndLoadFromMemory (config,  conf, FcTrue))   {
	// 	printf ("E: Unable to load config again\n");
	// 	ret = 4;
	// 	goto bail;
	//     }
	//     if (!FcConfigBuildFonts (config))  {
	// 	printf ("E: unable to build fonts again\n");
	// 	ret = 5;
	// 	goto bail;
	//     }
	//     fprintf (stderr, "D: Obtaining fonts information again\n");
	//     pat = FcPatternCreate ();
	//     fs = FcFontList (config, pat, NULL);
	//     FcPatternDestroy (pat);
	//     if (!fs || fs.nfont != 0)
	//     {
	// 	printf ("E: Unexpected the number of fonts: %d\n", !fs ? -1 : fs.nfont);
	// 	ret = 1;
	// 	goto bail;
	//     }
	//     FcFontSetDestroy (fs);
	//     fprintf (stderr, "D: Copying %s to %s\n", FONTFILE, fontdir);
	//     snprintf (cmd, 512, "sleep 1; cp -a %s %s; sleep 1", FONTFILE, fontdir);
	//  system (cmd);
	//     fprintf (stderr, "D: Reinitializing\n");
	//     if (FcConfigUptoDate(config))   {
	// 	fprintf (stderr, "E: Config up-to-date after addition\n");
	// 	ret = 3;
	// 	goto bail;
	//     }
	//     if (!FcInitReinitialize ())
	//     {
	// 	fprintf (stderr, "E: Unable to reinitialize\n");
	// 	ret = 2;
	// 	goto bail;
	//     }
	//     if (FcConfigGetCurrent () == config)
	//     {
	// 	fprintf (stderr, "E: config wasn't reloaded\n");
	// 	ret = 3;
	// 	goto bail;
	//     }
	//     FcConfigDestroy (config);

	//     config = FcConfigCreate ();
	//     if (!FcConfigParseAndLoadFromMemory (config,  conf, FcTrue))
	//     {
	// 	printf ("E: Unable to load config again\n");
	// 	ret = 4;
	// 	goto bail;
	//     }
	//     if (!FcConfigBuildFonts (config))
	//     {
	// 	printf ("E: unable to build fonts again\n");
	// 	ret = 5;
	// 	goto bail;
	//     }
	//     fprintf (stderr, "D: Obtaining fonts information\n");
	//     pat = FcPatternCreate ();
	//     fs = FcFontList (config, pat, NULL);
	//     FcPatternDestroy (pat);
	//     if (!fs || fs.nfont != 1)
	//     {
	// 	printf ("E: Unexpected the number of fonts: %d\n", !fs ? -1 : fs.nfont);
	// 	ret = 1;
	// 	goto bail;
	//     }
	//     FcFontSetDestroy (fs);
	//     FcConfigDestroy (config);

	// bail:
	//     fprintf (stderr, "Cleaning up\n");
	//     if (basedir)
	// 	unlink_dirs (basedir);
	//     if (fontdir)
	// 	FcStrFree (fontdir);
	//     if (cachedir)
	// 	FcStrFree (cachedir);

	//     return ret;

}
