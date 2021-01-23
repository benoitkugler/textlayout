package fontconfig

import (
	"bytes"
	"fmt"
)

// ported from fontconfig/src/fcinit.c Copyright © 2001 Keith Packard

const (
	CONFIGDIR        = "/usr/local/etc/fonts/conf.d"
	FC_CACHEDIR      = "/var/local/cache/fontconfig"
	FC_DEFAULT_FONTS = "<dir>/usr/share/fonts</dir>"
	FC_TEMPLATEDIR   = "/usr/local/share/fontconfig/conf.avail"
)

func initFallbackConfig() (*Config, error) {
	fallback := fmt.Sprintf(`	
	 <fontconfig>
	  	%s
		<cachedir>%s</cachedir>
		<cachedir prefix="xdg">fontconfig</cachedir>
		<include ignore_missing="yes">%s</include>
		<include ignore_missing="yes" prefix="xdg">fontconfig/conf.d</include>
		<include ignore_missing="yes" prefix="xdg">fontconfig/fonts.conf</include>
	 </fontconfig>
	 `, FC_DEFAULT_FONTS, FC_CACHEDIR, CONFIGDIR)

	config := NewConfig()

	err := config.LoadFromMemory(bytes.NewReader([]byte(fallback)))

	return config, err
}

// Load the configuration files
func initLoadOwnConfig() (*Config, error) {
	config := NewConfig()

	if err := config.parseConfig(""); err != nil {
		return initFallbackConfig()
	}

	err := config.parseConfig(FC_TEMPLATEDIR)
	if err != nil {
		return nil, err
	}

	// if len(config.cacheDirs) == 0 {
	// 	//  FcChar8 *prefix, *p;
	// 	//  size_t plen;
	// 	haveOwn := false

	// 	envFile := os.Getenv("FONTCONFIG_FILE")
	// 	envPath := os.Getenv("FONTCONFIG_PATH")
	// 	if envFile != "" || envPath != "" {
	// 		haveOwn = true
	// 	}

	// 	if !haveOwn {
	// 		fmt.Fprintf(os.Stderr, "fontconfig: no <cachedir> elements found. Check configuration.\n")
	// 		fmt.Fprintf(os.Stderr, "fontconfig: adding <cachedir>%s</cachedir>\n", FC_CACHEDIR)
	// 	}
	// 	prefix := xdgCacheHome()
	// 	if prefix == "" {
	// 		return initFallbackConfig(config.getSysRoot())
	// 	}
	// 	prefix = filepath.Join(prefix, "fontconfig")
	// 	if !haveOwn {
	// 		fmt.Fprintf(os.Stderr, "fontconfig: adding <cachedir prefix=\"xdg\">fontconfig</cachedir>\n")
	// 	}

	// 	err := config.addCacheDir(FC_CACHEDIR)
	// 	if err == nil {
	// 		err = config.addCacheDir(prefix)
	// 	}
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	return config, nil
}

//  FcConfig *
//  FcInitLoadConfig (void)
//  {
// 	 return initLoadOwnConfig (NULL);
//  }

// Loads the default configuration file and builds information about the
// available fonts.  Returns the resulting configuration.
func initLoadConfigAndFonts() (*Config, error) {
	config, err := initLoadOwnConfig()
	if err != nil {
		return nil, err
	}
	config.BuildFonts(nil) // TODO:
	return config, nil
}

//  /*
//   * Initialize the default library configuration
//   */
//  Bool
//  FcInit (void)
//  {
// 	 return FcConfigInit ();
//  }

//  /*
//   * Free all library-allocated data structures.
//   */
//  void
//  FcFini (void)
//  {
// 	 FcConfigFini ();
// 	 FcConfigPathFini ();
// 	 FcDefaultFini ();
// 	 FcObjectFini ();
// 	 FcCacheFini ();
//  }

//  /*
//   * Reread the configuration and available font lists
//   */
//  Bool
//  FcInitReinitialize (void)
//  {
// 	 FcConfig	*config;
// 	 Bool	ret;

// 	 config = initLoadConfigAndFonts ();
// 	 if (!config)
// 	 return FcFalse;
// 	 ret = FcConfigSetCurrent (config);
// 	 /* FcConfigSetCurrent() increases the refcount.
// 	  * decrease it here to avoid the memory leak.
// 	  */
// 	 FcConfigDestroy (config);

// 	 return ret;
//  }

//  Bool
//  FcInitBringUptoDate (void)
//  {
// 	 FcConfig	*config = FcConfigReference (NULL);
// 	 Bool	ret = FcTrue;
// 	 time_t	now;

// 	 if (!config)
// 	 return FcFalse;
// 	 /*
// 	  * rescanInterval == 0 disables automatic up to date
// 	  */
// 	 if (config.rescanInterval == 0)
// 	 goto bail;
// 	 /*
// 	  * Check no more often than rescanInterval seconds
// 	  */
// 	 now = time (0);
// 	 if (config.rescanTime + config.rescanInterval - now > 0)
// 	 goto bail;
// 	 /*
// 	  * If up to date, don't reload configuration
// 	  */
// 	 if (FcConfigUptoDate (0))
// 	 goto bail;
// 	 ret = FcInitReinitialize ();
//  bail:
// 	 FcConfigDestroy (config);

// 	 return ret;
//  }
