package fontconfig

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// ported from fontconfig: Copyright 2000 Keith Packard, 2005 Patrick Lam

const cacheSuffix = ".cache-0"

// for now we dont build an advanced cache footprint:
// we simply use the directory
// if a directory is modified, it then must be entirely scanned
type FcCache string

/// Build a cache structure from the given contents
func FcDirCacheBuild(_ FcFontSet, dir string, _ FcStrSet) FcCache {
	return FcCache(dir)
}

// Read (or construct) the cache for a directory
func FcDirCacheRead(dir string, force bool, config *FcConfig) FcCache {
	var cache FcCache

	config = fallbackConfig(config)
	// Try to use existing cache file
	if !force {
		cache, _ = FcDirCacheLoad(dir, config)
	}

	// Not using existing cache file, construct new cache
	if cache == "" {
		cache = FcDirCacheScan(dir, config)
	}

	return cache
}

//  Scan the specified directory and construct a cache of its contents
func FcDirCacheScan(dir string, config *FcConfig) FcCache {
	//  FcStrSet		*dirs;
	//  FcFontSet		*set;
	//  FcCache		*cache = NULL;
	//  struct stat		dirStat;
	sysroot := config.getSysRoot()
	//  #ifndef _WIN32
	// 	 int			fd = -1;
	//  #endif

	d := dir
	if sysroot != "" {
		d = filepath.Join(sysroot, dir)
	}

	if debugMode {
		fmt.Printf("cache scan dir %s\n", d)
	}

	_, err := os.Stat(d)
	if err != nil {
		return ""
	}

	var (
		set  FcFontSet
		dirs = make(FcStrSet)
	)

	//  #ifndef _WIN32
	// 	 fd = FcDirCacheLock (dir, config);
	//  #endif
	// Scan the dir

	// Do not pass sysroot here. FcDirScanConfig() do take care of it
	if !FcDirScanConfig(&set, dirs, dir, config) {
		return ""
	}

	// Build the cache object
	cache := FcDirCacheBuild(set, dir, dirs)

	/// Write out the cache file, ignoring any troubles
	FcDirCacheWrite(cache, config)

	//  #ifndef _WIN32
	// 	 FcDirCacheUnlock (fd);
	//  #endif

	return cache
}

func FcDirScanConfig(set *FcFontSet, dirs FcStrSet, dir string, config *FcConfig) bool {
	sysroot := config.getSysRoot()

	sDir := dir
	if sysroot != "" {
		sDir = filepath.Join(sysroot, dir)
	}

	if debugMode {
		fmt.Printf("\tScanning dir %s\n", sDir)
	}

	filesList, err := ioutil.ReadDir(sDir)
	if err != nil {
		/* Don't complain about missing directories */
		return false
	}

	// Scan file files to build font patterns
	for _, e := range filesList {
		name := e.Name()
		if name[0] == '.' {
			continue
		}
		FcFileScanConfig(set, dirs, filepath.Join(dir, name), config)
	}

	return true
}

func FcDirCacheLoad(dir string, config *FcConfig) (FcCache, string) {
	var cache FcCache

	// config = fallbackConfig(config)

	// if !FcDirCacheProcess(config, dir,
	// 	FcDirCacheMapHelper,
	// 	&cache, cache_file) {
	// 	cache = nil
	// }

	return cache, ""
}

// type callbackProcess = func(config *FcConfig, int fd, struct stat *fdStat,  struct stat *dirStat, struct timeval *cacheMtime) bool
type callbackProcess = func(config *FcConfig, fd *os.File, dirStat os.FileInfo) *FcCache

// Look for a cache file for the specified dir. Attempt
// to use each one we find, stopping when the callback
// indicates success
func FcDirCacheProcess(config *FcConfig, dir string, callback callbackProcess) (*FcCache, string) {
	// int		fd = -1;
	// FcChar8	cacheBase[CACHEBASE_LEN];
	// FcStrList	*list;
	// FcChar8	*cacheDir, *d;
	// struct stat file_stat, dirStat;
	// FcBool	ret = FcFalse;
	sysroot := config.getSysRoot()
	// struct timeval latest_mtime = (struct timeval){ 0 };

	d := dir
	if sysroot != "" {
		d = filepath.Join(sysroot, dir)
	}

	dirstat, err := os.Stat(d)
	if err != nil {
		return nil, ""
	}

	cacheBase := FcDirCacheBasenameMD5(config, dir)
	var cacheFileRet string
	for cacheDir := range config.cacheDirs {
		var cacheHashed string

		if sysroot != "" {
			cacheHashed = filepath.Join(sysroot, cacheDir, cacheBase)
		} else {
			cacheHashed = filepath.Join(cacheDir, cacheBase)
		}

		fd, err := os.Open(cacheHashed)
		if err != nil {
			continue
		}

		v := callback(config, fd, dirstat)
		_ = fd.Close()
		if v != nil {
			cacheFileRet = cacheHashed
		}

	}

	return nil, cacheFileRet
}

// FcBool
// FcDirCacheCreateUUID (FcChar8  *dir,
// 		      FcBool    force,
// 		      config *FcConfig)
// {
//     return FcTrue;
// }

// FcBool
// FcDirCacheDeleteUUID (const FcChar8  *dir,
// 		      FcConfig       *config)
// {
//     FcBool ret = FcTrue;
// #ifndef _WIN32
//     const FcChar8 *sysroot;
//     FcChar8 *target, *d;
//     struct stat statb;
//     struct timeval times[2];

//     config = FcConfigReference (config);
//     if (!config)
// 	return FcFalse;
//     sysroot = config.getSysRoot ();
//     if (sysroot)
// 	d = FcStrBuildFilename (sysroot, dir, nil);
//     else
// 	d = FcStrBuildFilename (dir, nil);
//     if (FcStat (d, &statb) != 0)
//     {
// 	ret = FcFalse;
// 	goto bail;
//     }
//     target = FcStrBuildFilename (d, ".uuid", nil);
//     ret = unlink ((char *) target) == 0;
//     if (ret)
//     {
// 	times[0].tv_sec = statb.st_atime;
// 	times[1].tv_sec = statb.st_mtime;
// #ifdef HAVE_STRUCT_STAT_ST_MTIM
// 	times[0].tv_usec = statb.st_atim.tv_nsec / 1000;
// 	times[1].tv_usec = statb.st_mtim.tv_nsec / 1000;
// #else
// 	times[0].tv_usec = 0;
// 	times[1].tv_usec = 0;
// #endif
// 	if (utimes ((const char *) d, times) != 0)
// 	{
// 	    fprintf (stderr, "Unable to revert mtime: %s\n", d);
// 	}
//     }
//     FcStrFree (target);
// bail:
//     FcStrFree (d);
// #endif
//     FcConfigDestroy (config);

//     return ret;
// }

// #define CACHEBASE_LEN (1 + 36 + 1 + sizeof (FC_ARCHITECTURE) + sizeof (FC_CACHE_SUFFIX))

// static FcBool
// FcCacheIsMmapSafe (int fd)
// {
//     enum {
//       MMAP_NOT_INITIALIZED = 0,
//       MMAP_USE,
//       MMAP_DONT_USE,
//       MMAP_CHECK_FS,
//     } status;
//     static void *static_status;

//     status = (intptr_t) fc_atomic_ptr_get (&static_status);

//     if (status == MMAP_NOT_INITIALIZED)
//     {
// 	const char *env = getenv ("FONTCONFIG_USE_MMAP");
// 	FcBool use;
// 	if (env && FcNameBool ((const FcChar8 *) env, &use))
// 	    status =  use ? MMAP_USE : MMAP_DONT_USE;
// 	else
// 	    status = MMAP_CHECK_FS;
// 	(void) fc_atomic_ptr_cmpexch (&static_status, nil, (void *) status);
//     }

//     if (status == MMAP_CHECK_FS)
// 	return FcIsFsMmapSafe (fd);
//     else
// 	return status == MMAP_USE;

// }

func (config *FcConfig) mapFontPath(path string) string {
	var dir string
	for dir = range config.fontDirs {
		if strings.HasPrefix(dir, path) {
			break
		}
	}
	return dir
}

func FcDirCacheBasenameMD5(config *FcConfig, dir string) string {
	var (
		origDir string
		saltB   = make([]byte, 16)
	)
	rand.Read(saltB)
	salt := hex.EncodeToString(saltB)

	/* Obtain a path where "dir" is mapped to.
	 * In case:
	 * <remap-dir as-path="/usr/share/fonts">/run/host/fonts</remap-dir>
	 *
	 * mapFontPath (config, "/run/host/fonts") will returns "/usr/share/fonts".
	 */
	mappedDir := config.mapFontPath(dir)
	if mappedDir != "" {
		origDir = dir
		dir = mappedDir
	}
	if salt != "" {
		if origDir == "" {
			origDir = dir
		}
		dir = dir + salt
	}

	hash := md5.Sum([]byte(dir))

	hexHash := hex.EncodeToString(hash[:])
	cacheBase := "/" + hexHash + runtime.GOOS + runtime.GOARCH + cacheSuffix

	if debugMode {
		fmt.Printf("cache: %s (dir: %s %s %s)\n", cacheBase, origDir, mappedDir, salt)
	}

	return cacheBase
}

func FcDirCacheMapHelper(config *FcConfig, fd *os.File, fdStat, dirStat os.FileInfo) (*FcCache, int64) {
	cache := cacheMapFd(config, fd, fdStat, dirStat)
	// struct timeval cacheMtime, zeroMtime = { 0, 0}, dirMtime;

	// if (!cache)
	// return FcFalse;
	cacheMtime := fdStat.ModTime().UnixNano()
	// dirMtime := dirStat.ModTime().UnixNano()
	var latestCacheMtime int64
	/* special take care of OSTree */
	// if (!timercmp (&zeroMtime, &dirMtime, !=)){
	// if (!timercmp (&zeroMtime, &cacheMtime, !=)){
	//     if (*((FcCache **) closure))
	// 	FcDirCacheUnload (*((FcCache **) closure));
	// } else if (*((FcCache **) closure) && !timercmp (&zeroMtime, latestCacheMtime, !=)) 	{
	//     FcDirCacheUnload (cache);
	//     return FcFalse;
	// } else if (timercmp (latestCacheMtime, &cacheMtime, <)) 	{
	//     if (*((FcCache **) closure)){
	// 	FcDirCacheUnload (*((FcCache **) closure));}
	// }
	// }  else if (timercmp (latestCacheMtime, &cacheMtime, <))    {
	// if (*((FcCache **) closure)){
	//     FcDirCacheUnload (*((FcCache **) closure));}
	// }  else   {
	// FcDirCacheUnload (cache);
	// return FcFalse;
	// }

	latestCacheMtime = cacheMtime
	return cache, latestCacheMtime
}

// Map a cache file into memory
func cacheMapFd(config *FcConfig, fd *os.File, fdStat, dirStat os.FileInfo) *FcCache {
	// TODO:

	cache := FcCacheFindByStat(fdStat)
	// if cache != nil {
	// 	if FcCacheTimeValid(config, cache, dirStat) {
	// 		return cache
	// 	}
	// 	FcDirCacheUnload(cache)
	// 	cache = nil
	// }

	// if !cache {
	// 	cache = malloc(fdStat.st_size)
	// 	if !cache {
	// 		return nil
	// 	}

	// 	if read(fd, cache, fdStat.st_size) != fdStat.st_size {
	// 		free(cache)
	// 		return nil
	// 	}
	// 	allocated = FcTrue
	// }
	// if cache.magic != FC_CACHE_MAGIC_MMAP ||
	// 	cache.version < FC_CACHE_VERSION_NUMBER ||
	// 	cache.size != fdStat.st_size ||
	// 	!FcCacheOffsetsValid(cache) ||
	// 	!FcCacheTimeValid(config, cache, dirStat) ||
	// 	!FcCacheInsert(cache, fdStat) {

	// 	return nil
	// }

	return cache
}

// #ifndef _WIN32
// static FcChar8 *
// FcDirCacheBasenameUUID (config *FcConfig, dir string, FcChar8 cacheBase[CACHEBASE_LEN])
// {
//     FcChar8 *target, *fuuid;
//     const FcChar8 *sysroot = config.getSysRoot ();
//     int fd;

//     /* We don't need to apply remapping here. because .uuid was created at that very directory
//      * to determine the cache name no matter where it was mapped to.
//      */
//     cacheBase[0] = 0;
//     if (sysroot)
// 	target = FcStrBuildFilename (sysroot, dir, nil);
//     else
// 	target = FcStrdup (dir);
//     fuuid = FcStrBuildFilename (target, ".uuid", nil);
//     if ((fd = FcOpen ((char *) fuuid, O_RDONLY)) != -1)
//     {
// 	char suuid[37];
// 	ssize_t len;

// 	memset (suuid, 0, sizeof (suuid));
// 	len = read (fd, suuid, 36);
// 	suuid[36] = 0;
// 	close (fd);
// 	if (len < 0)
// 	    goto bail;
// 	cacheBase[0] = '/';
// 	strcpy ((char *)&cacheBase[1], suuid);
// 	strcat ((char *) cacheBase, "-" FC_ARCHITECTURE FC_CACHE_SUFFIX);
// 	if (FcDebug () & FC_DBG_CACHE)
// 	{
// 	    printf ("cache fallbacks to: %s (dir: %s)\n", cacheBase, dir);
// 	}
//     }
// bail:
//     FcStrFree (fuuid);
//     FcStrFree (target);

//     return cacheBase;
// }
// #endif

// FcBool
// FcDirCacheUnlink (dir string, config *FcConfig)
// {
//     FcChar8	*cacheHashed = nil;
//     FcChar8	cacheBase[CACHEBASE_LEN];
// #ifndef _WIN32
//     FcChar8     uuid_cacheBase[CACHEBASE_LEN];
// #endif
//     FcStrList	*list;
//     FcChar8	*cacheDir;
//     const FcChar8 *sysroot;
//     FcBool	ret = FcTrue;

//     config = FcConfigReference (config);
//     if (!config)
// 	return FcFalse;
//     sysroot = config.getSysRoot ();

//     FcDirCacheBasenameMD5 (config, dir, cacheBase);
// #ifndef _WIN32
//     FcDirCacheBasenameUUID (config, dir, uuid_cacheBase);
// #endif

//     list = FcStrListCreate (config.cacheDirs);
//     if (!list)
//     {
// 	ret = FcFalse;
// 	goto bail;
//     }

//     while ((cacheDir = FcStrListNext (list)))
//     {
// 	if (sysroot)
// 	    cacheHashed = FcStrBuildFilename (sysroot, cacheDir, cacheBase, nil);
// 	else
// 	    cacheHashed = FcStrBuildFilename (cacheDir, cacheBase, nil);
//         if (!cacheHashed)
// 	    break;
// 	(void) unlink ((char *) cacheHashed);
// 	FcStrFree (cacheHashed);
// #ifndef _WIN32
// 	if (uuid_cacheBase[0] != 0)
// 	{
// 	    if (sysroot)
// 		cacheHashed = FcStrBuildFilename (sysroot, cacheDir, uuid_cacheBase, nil);
// 	    else
// 		cacheHashed = FcStrBuildFilename (cacheDir, uuid_cacheBase, nil);
// 	    if (!cacheHashed)
// 		break;
// 	    (void) unlink ((char *) cacheHashed);
// 	    FcStrFree (cacheHashed);
// 	}
// #endif
//     }
//     FcStrListDone (list);
//     FcDirCacheDeleteUUID (dir, config);
//     /* return FcFalse if something went wrong */
//     if (cacheDir)
// 	ret = FcFalse;
// bail:
//     FcConfigDestroy (config);

//     return ret;
// }

// static int
// FcDirCacheOpenFile (const FcChar8 *cache_file, struct stat *file_stat)
// {
//     int	fd;

// #ifdef _WIN32
//     if (FcStat (cache_file, file_stat) < 0)
//         return -1;
// #endif
//     fd = FcOpen((char *) cache_file, O_RDONLY | O_BINARY);
//     if (fd < 0)
// 	return fd;
// #ifndef _WIN32
//     if (fstat (fd, file_stat) < 0)
//     {
// 	close (fd);
// 	return -1;
//     }
// #endif
//     return fd;
// }

// #define FC_CACHE_MIN_MMAP   1024

// /*
//  * Skip list element, make sure the 'next' pointer is the last thing
//  * in the structure, it will be allocated large enough to hold all
//  * of the necessary pointers
//  */

// typedef struct _FcCacheSkip FcCacheSkip;

type FcCacheSkip struct {
	//     FcCache	    *cache;
	//     FcRef	    ref;
	//     intptr_t	    size;
	//     void	   *allocated;
	//     dev_t	    cache_dev;
	//     ino_t	    cache_ino;
	//     time_t	    cacheMtime;
	//     long	    cacheMtime_nano;
	//     FcCacheSkip	    *next[1];
}

// /*
//  * The head of the skip list; pointers for every possible level
//  * in the skip list, plus the largest level in the list
//  */

const FC_CACHE_MAX_LEVEL = 16

/* Protected by cacheLock below */
var (
	fcCacheChains [FC_CACHE_MAX_LEVEL][]FcCacheSkip
	cacheLock     sync.Mutex
)

// static int		fcCacheMaxLevel;

// static void
// free_lock (void)
// {
//   FcMutex *lock;
//   lock = fc_atomic_ptr_get (&cacheLock);
//   if (lock && fc_atomic_ptr_cmpexch (&cacheLock, lock, nil)) {
//     FcMutexFinish (lock);
//     free (lock);
//   }
// }

// /*
//  * Generate a random level number, distributed
//  * so that each level is 1/4 as likely as the one before
//  *
//  * Note that level numbers run 1 <= level <= MAX_LEVEL
//  */
// static int
// random_level (void)
// {
//     /* tricky bit -- each bit is '1' 75% of the time */
//     long int	bits = FcRandom () | FcRandom ();
//     int	level = 0;

//     while (++level < FC_CACHE_MAX_LEVEL)
//     {
// 	if (bits & 1)
// 	    break;
// 	bits >>= 1;
//     }
//     return level;
// }

// /*
//  * Insert cache into the list
//  */
// static FcBool
// FcCacheInsert (FcCache *cache, struct stat *cacheStat)
// {
//     FcCacheSkip    **update[FC_CACHE_MAX_LEVEL];
//     FcCacheSkip    *s, **next;
//     int		    i, level;

//     lock_cache ();

//     /*
//      * Find links along each chain
//      */
//     next = fcCacheChains;
//     for (i = fcCacheMaxLevel; --i >= 0; )
//     {
// 	for (; (s = next[i]); next = s.next)
// 	    if (s.cache > cache)
// 		break;
//         update[i] = &next[i];
//     }

//     /*
//      * Create new list element
//      */
//     level = random_level ();
//     if (level > fcCacheMaxLevel)
//     {
// 	level = fcCacheMaxLevel + 1;
// 	update[fcCacheMaxLevel] = &fcCacheChains[fcCacheMaxLevel];
// 	fcCacheMaxLevel = level;
//     }

//     s = malloc (sizeof (FcCacheSkip) + (level - 1) * sizeof (FcCacheSkip *));
//     if (!s)
// 	return FcFalse;

//     s.cache = cache;
//     s.size = cache.size;
//     s.allocated = nil;
//     FcRefInit (&s.ref, 1);
//     if (cacheStat)
//     {
// 	s.cache_dev = cacheStat.st_dev;
// 	s.cache_ino = cacheStat.st_ino;
// 	s.cacheMtime = cacheStat.st_mtime;
// #ifdef HAVE_STRUCT_STAT_ST_MTIM
// 	s.cacheMtime_nano = cacheStat.st_mtim.tv_nsec;
// #else
// 	s.cacheMtime_nano = 0;
// #endif
//     }
//     else
//     {
// 	s.cache_dev = 0;
// 	s.cache_ino = 0;
// 	s.cacheMtime = 0;
// 	s.cacheMtime_nano = 0;
//     }

//     /*
//      * Insert into all fcCacheChains
//      */
//     for (i = 0; i < level; i++)
//     {
// 	s.next[i] = *update[i];
// 	*update[i] = s;
//     }

//     unlock_cache ();
//     return FcTrue;
// }

// static FcCacheSkip *
// FcCacheFindByAddrUnlocked (void *object)
// {
//     int	    i;
//     FcCacheSkip    **next = fcCacheChains;
//     FcCacheSkip    *s;

//     if (!object)
// 	return nil;

//     /*
//      * Walk chain pointers one level at a time
//      */
//     for (i = fcCacheMaxLevel; --i >= 0;)
// 	while (next[i] && (char *) object >= ((char *) next[i].cache + next[i].size))
// 	    next = next[i].next;
//     /*
//      * Here we are
//      */
//     s = next[0];
//     if (s && (char *) object < ((char *) s.cache + s.size))
// 	return s;
//     return nil;
// }

// static FcCacheSkip *
// FcCacheFindByAddr (void *object)
// {
//     FcCacheSkip *ret;
//     lock_cache ();
//     ret = FcCacheFindByAddrUnlocked (object);
//     unlock_cache ();
//     return ret;
// }

// static void
// FcCacheRemoveUnlocked (FcCache *cache)
// {
//     FcCacheSkip	    **update[FC_CACHE_MAX_LEVEL];
//     FcCacheSkip	    *s, **next;
//     int		    i;
//     void            *allocated;

//     /*
//      * Find links along each chain
//      */
//     next = fcCacheChains;
//     for (i = fcCacheMaxLevel; --i >= 0; )
//     {
// 	for (; (s = next[i]); next = s.next)
// 	    if (s.cache >= cache)
// 		break;
//         update[i] = &next[i];
//     }
//     s = next[0];
//     for (i = 0; i < fcCacheMaxLevel && *update[i] == s; i++)
// 	*update[i] = s.next[i];
//     while (fcCacheMaxLevel > 0 && fcCacheChains[fcCacheMaxLevel - 1] == nil)
// 	fcCacheMaxLevel--;

//     if (s)
//     {
// 	allocated = s.allocated;
// 	while (allocated)
// 	{
// 	    /* First element in allocated chunk is the free list */
// 	    next = *(void **)allocated;
// 	    free (allocated);
// 	    allocated = next;
// 	}
// 	free (s);
//     }
// }

func FcCacheFindByStat(cacheStat os.FileInfo) *FcCache {
	// TODO:
	// FcCacheSkip * s
	cacheLock.Lock()
	defer cacheLock.Unlock()

	// for _, s := range fcCacheChains[0] {
	// 	if s.cache_dev == cacheStat.st_dev &&
	// 		s.cache_ino == cacheStat.st_ino &&
	// 		s.cacheMtime == cacheStat.st_mtime {
	// 		// #ifdef HAVE_STRUCT_STAT_ST_MTIM
	// 		// 	    if (s.cacheMtime_nano != cacheStat.st_mtim.tv_nsec)
	// 		// 		continue;
	// 		// #endif
	// 		FcRefInc(&s.ref)
	// 		unlock_cache()
	// 		return s.cache
	// 	}
	// }
	// unlock_cache()
	return nil
}

// static void
// FcDirCacheDisposeUnlocked (FcCache *cache)
// {
//     FcCacheRemoveUnlocked (cache);

//     switch (cache.magic) {
//     case FC_CACHE_MAGIC_ALLOC:
// 	free (cache);
// 	break;
//     case FC_CACHE_MAGIC_MMAP:
// #if defined(HAVE_MMAP) || defined(__CYGWIN__)
// 	munmap (cache, cache.size);
// #elif defined(_WIN32)
// 	UnmapViewOfFile (cache);
// #endif
// 	break;
//     }
// }

// void
// FcCacheObjectReference (void *object)
// {
//     FcCacheSkip *skip = FcCacheFindByAddr (object);

//     if (skip)
// 	FcRefInc (&skip.ref);
// }

// void
// FcCacheObjectDereference (void *object)
// {
//     FcCacheSkip	*skip;

//     lock_cache ();
//     skip = FcCacheFindByAddrUnlocked (object);
//     if (skip)
//     {
// 	if (FcRefDec (&skip.ref) == 1)
// 	    FcDirCacheDisposeUnlocked (skip.cache);
//     }
//     unlock_cache ();
// }

// void *
// FcCacheAllocate (FcCache *cache, size_t len)
// {
//     FcCacheSkip	*skip;
//     void *allocated = nil;

//     lock_cache ();
//     skip = FcCacheFindByAddrUnlocked (cache);
//     if (skip)
//     {
//       void *chunk = malloc (sizeof (void *) + len);
//       if (chunk)
//       {
// 	  /* First element in allocated chunk is the free list */
// 	  *(void **)chunk = skip.allocated;
// 	  skip.allocated = chunk;
// 	  /* Return the rest */
// 	  allocated = ((FcChar8 *)chunk) + sizeof (void *);
//       }
//     }
//     unlock_cache ();
//     return allocated;
// }

// void
// FcCacheFini (void)
// {
//     int		    i;

//     for (i = 0; i < FC_CACHE_MAX_LEVEL; i++)
//     {
// 	if (FcDebug() & FC_DBG_CACHE)
// 	{
// 	    if (fcCacheChains[i] != nil)
// 	    {
// 		FcCacheSkip *s = fcCacheChains[i];
// 		printf("Fontconfig error: not freed %p (dir: %s, refcount %" FC_ATOMIC_INT_FORMAT ")\n", s.cache, FcCacheDir(s.cache), s.ref.count);
// 	    }
// 	}
// 	else
// 	    assert (fcCacheChains[i] == nil);
//     }
//     assert (fcCacheMaxLevel == 0);

//     free_lock ();
// }

// static FcBool
// FcCacheTimeValid (config *FcConfig, FcCache *cache, struct stat *dirStat)
// {
//     struct stat	dirStatic;
//     FcBool fnano = FcTrue;

//     if (!dirStat)
//     {
// 	const FcChar8 *sysroot = config.getSysRoot ();
// 	FcChar8 *d;

// 	if (sysroot)
// 	    d = FcStrBuildFilename (sysroot, FcCacheDir (cache), nil);
// 	else
// 	    d = FcStrdup (FcCacheDir (cache));
// 	if (FcStatChecksum (d, &dirStatic) < 0)
// 	{
// 	    FcStrFree (d);
// 	    return FcFalse;
// 	}
// 	FcStrFree (d);
// 	dirStat = &dirStatic;
//     }
// #ifdef HAVE_STRUCT_STAT_ST_MTIM
//     fnano = (cache.checksum_nano == dirStat.st_mtim.tv_nsec);
//     if (FcDebug () & FC_DBG_CACHE)
// 	printf ("FcCacheTimeValid dir \"%s\" cache checksum %d.%ld dir checksum %d.%ld\n",
// 		FcCacheDir (cache), cache.checksum, (long)cache.checksum_nano, (int) dirStat.st_mtime, dirStat.st_mtim.tv_nsec);
// #else
//     if (FcDebug () & FC_DBG_CACHE)
// 	printf ("FcCacheTimeValid dir \"%s\" cache checksum %d dir checksum %d\n",
// 		FcCacheDir (cache), cache.checksum, (int) dirStat.st_mtime);
// #endif

//     return dirStat.st_mtime == 0 || (cache.checksum == (int) dirStat.st_mtime && fnano);
// }

// static FcBool
// FcCacheOffsetsValid (FcCache *cache)
// {
//     char		*base = (char *)cache;
//     char		*end = base + cache.size;
//     intptr_t		*dirs;
//     FcFontSet		*fs;
//     int			 i, j;

//     if (cache.dir < 0 || cache.dir > cache.size - sizeof (intptr_t) ||
//         memchr (base + cache.dir, '\0', cache.size - cache.dir) == nil)
//         return FcFalse;

//     if (cache.dirs < 0 || cache.dirs >= cache.size ||
//         cache.dirs_count < 0 ||
//         cache.dirs_count > (cache.size - cache.dirs) / sizeof (intptr_t))
//         return FcFalse;

//     dirs = FcCacheDirs (cache);
//     if (dirs)
//     {
//         for (i = 0; i < cache.dirs_count; i++)
//         {
//             FcChar8	*dir;

//             if (dirs[i] < 0 ||
//                 dirs[i] > end - (char *) dirs - sizeof (intptr_t))
//                 return FcFalse;

//             dir = FcOffsetToPtr (dirs, dirs[i], FcChar8);
//             if (memchr (dir, '\0', end - (char *) dir) == nil)
//                 return FcFalse;
//          }
//     }

//     if (cache.set < 0 || cache.set > cache.size - sizeof (FcFontSet))
//         return FcFalse;

//     fs = FcCacheSet (cache);
//     if (fs)
//     {
//         if (fs.nfont > (end - (char *) fs) / sizeof (FcPattern))
//             return FcFalse;

//         if (!FcIsEncodedOffset(fs.fonts))
//             return FcFalse;

//         for (i = 0; i < fs.nfont; i++)
//         {
//             FcPattern		*font = FcFontSetFont (fs, i);
//             FcPatternElt	*e;
//             FcValueListPtr	 l;
// 	    char                *last_offset;

//             if ((char *) font < base ||
//                 (char *) font > end - sizeof (FcFontSet) ||
//                 font.elts_offset < 0 ||
//                 font.elts_offset > end - (char *) font ||
//                 font.num > (end - (char *) font - font.elts_offset) / sizeof (FcPatternElt) ||
// 		!FcRefIsConst (&font.ref))
//                 return FcFalse;

//             e = FcPatternElts(font);
//             if (e.values != 0 && !FcIsEncodedOffset(e.values))
//                 return FcFalse;

// 	    for (j = 0; j < font.num; j++)
// 	    {
// 		last_offset = (char *) font + font.elts_offset;
// 		for (l = FcPatternEltValues(&e[j]); l; l = FcValueListNext(l))
// 		{
// 		    if ((char *) l < last_offset || (char *) l > end - sizeof (*l) ||
// 			(l.next != nil && !FcIsEncodedOffset(l.next)))
// 			return FcFalse;
// 		    last_offset = (char *) l + 1;
// 		}
// 	    }
//         }
//     }

//     return FcTrue;
// }

// void
// FcDirCacheReference (FcCache *cache, int nref)
// {
//     FcCacheSkip *skip = FcCacheFindByAddr (cache);

//     if (skip)
// 	FcRefAdd (&skip.ref, nref);
// }

// void
// FcDirCacheUnload (FcCache *cache)
// {
//     FcCacheObjectDereference (cache);
// }

// FcCache *
// FcDirCacheLoadFile (const FcChar8 *cache_file, struct stat *file_stat)
// {
//     int	fd;
//     FcCache *cache;
//     struct stat	my_file_stat;
//     config *FcConfig;

//     if (!file_stat)
// 	file_stat = &my_file_stat;
//     config = FcConfigReference (nil);
//     if (!config)
// 	return nil;
//     fd = FcDirCacheOpenFile (cache_file, file_stat);
//     if (fd < 0)
// 	return nil;
//     cache = cacheMapFd (config, fd, file_stat, nil);
//     FcConfigDestroy (config);
//     close (fd);
//     return cache;
// }

// static int
// FcDirChecksum (struct stat *statb)
// {
//     int			ret = (int) statb.st_mtime;
//     char		*endptr;
//     char		*source_date_epoch;
//     unsigned long long	epoch;

//     source_date_epoch = getenv("SOURCE_DATE_EPOCH");
//     if (source_date_epoch)
//     {
// 	errno = 0;
// 	epoch = strtoull(source_date_epoch, &endptr, 10);

// 	if (endptr == source_date_epoch)
// 	    fprintf (stderr,
// 		     "Fontconfig: SOURCE_DATE_EPOCH invalid\n");
// 	else if ((errno == ERANGE && (epoch == ULLONG_MAX || epoch == 0))
// 		|| (errno != 0 && epoch == 0))
// 	    fprintf (stderr,
// 		     "Fontconfig: SOURCE_DATE_EPOCH: strtoull: %s: %" FC_UINT64_FORMAT "\n",
// 		     strerror(errno), epoch);
// 	else if (*endptr != '\0')
// 	    fprintf (stderr,
// 		     "Fontconfig: SOURCE_DATE_EPOCH has trailing garbage\n");
// 	else if (epoch > ULONG_MAX)
// 	    fprintf (stderr,
// 		     "Fontconfig: SOURCE_DATE_EPOCH must be <= %lu but saw: %" FC_UINT64_FORMAT "\n",
// 		     ULONG_MAX, epoch);
// 	else if (epoch < ret)
// 	    /* Only override if directory is newer */
// 	    ret = (int) epoch;
//     }

//     return ret;
// }

// static int64_t
// FcDirChecksumNano (struct stat *statb)
// {
// #ifdef HAVE_STRUCT_STAT_ST_MTIM
//     /* No nanosecond component to parse */
//     if (getenv("SOURCE_DATE_EPOCH"))
// 	return 0;
//     return statb.st_mtim.tv_nsec;
// #else
//     return 0;
// #endif
// }

// /*
//  * Validate a cache file by reading the header and checking
//  * the magic number and the size field
//  */
// static FcBool
// FcDirCacheValidateHelper (config *FcConfig, int fd, struct stat *fdStat, struct stat *dirStat, struct timeval *latestCacheMtime, void *closure FC_UNUSED)
// {
//     FcBool  ret = FcTrue;
//     FcCache	c;

//     if (read (fd, &c, sizeof (FcCache)) != sizeof (FcCache))
// 	ret = FcFalse;
//     else if (c.magic != FC_CACHE_MAGIC_MMAP)
// 	ret = FcFalse;
//     else if (c.version < FC_CACHE_VERSION_NUMBER)
// 	ret = FcFalse;
//     else if (fdStat.st_size != c.size)
// 	ret = FcFalse;
//     else if (c.checksum != FcDirChecksum (dirStat))
// 	ret = FcFalse;
// #ifdef HAVE_STRUCT_STAT_ST_MTIM
//     else if (c.checksum_nano != FcDirChecksumNano (dirStat))
// 	ret = FcFalse;
// #endif
//     return ret;
// }

// static FcBool
// FcDirCacheValidConfig (dir string, config *FcConfig)
// {
//     return FcDirCacheProcess (config, dir,
// 			      FcDirCacheValidateHelper,
// 			      nil, nil);
// }

// FcBool
// FcDirCacheValid (dir string)
// {
//     FcConfig	*config;
//     FcBool	ret;

//     config = FcConfigReference (nil);
//     if (!config)
//         return FcFalse;

//     ret = FcDirCacheValidConfig (dir, config);
//     FcConfigDestroy (config);

//     return ret;
// }

// FcCache *
// FcDirCacheRebuild (FcCache *cache, struct stat *dirStat, FcStrSet *dirs)
// {
//     FcCache *new;
//     FcFontSet *set = FcFontSetDeserialize (FcCacheSet (cache));
//     dir string = FcCacheDir (cache);

//     new = FcDirCacheBuild (set, dir, dirStat, dirs);
//     FcFontSetDestroy (set);

//     return new;
// }

/* write serialized state to the cache file */
func FcDirCacheWrite(cache FcCache, config *FcConfig) {
	// TODO:
	//     FcChar8	    *dir = FcCacheDir (cache);
	//     FcChar8	    cacheBase[CACHEBASE_LEN];
	//     FcChar8	    *cacheHashed;
	//     int 	    fd;
	//     FcAtomic 	    *atomic;
	//     FcStrList	    *list;
	//     FcChar8	    *cacheDir = nil;
	//     FcChar8	    *test_dir, *d = nil;
	//     FcCacheSkip     *skip;
	//     struct stat     cacheStat;
	//     unsigned int    magic;
	//     int		    written;
	//     const FcChar8   *sysroot = config.getSysRoot ();

	//     /*
	//      * Write it to the first directory in the list which is writable
	//      */

	//     list = FcStrListCreate (config.cacheDirs);
	//     if (!list)
	// 	return FcFalse;
	//     while ((test_dir = FcStrListNext (list)))
	//     {
	// 	if (d)
	// 	    FcStrFree (d);
	// 	if (sysroot)
	// 	    d = FcStrBuildFilename (sysroot, test_dir, nil);
	// 	else
	// 	    d = FcStrCopyFilename (test_dir);

	// 	if (access ((char *) d, W_OK) == 0)
	// 	{
	// 	    cacheDir = FcStrCopyFilename (d);
	// 	    break;
	// 	}
	// 	else
	// 	{
	// 	    /*
	// 	     * If the directory doesn't exist, try to create it
	// 	     */
	// 	    if (access ((char *) d, F_OK) == -1) {
	// 		if (FcMakeDirectory (d))
	// 		{
	// 		    cacheDir = FcStrCopyFilename (d);
	// 		    /* Create CACHEDIR.TAG */
	// 		    FcDirCacheCreateTagFile (d);
	// 		    break;
	// 		}
	// 	    }
	// 	    /*
	// 	     * Otherwise, try making it writable
	// 	     */
	// 	    else if (chmod ((char *) d, 0755) == 0)
	// 	    {
	// 		cacheDir = FcStrCopyFilename (d);
	// 		/* Try to create CACHEDIR.TAG too */
	// 		FcDirCacheCreateTagFile (d);
	// 		break;
	// 	    }
	// 	}
	//     }
	//     if (!test_dir)
	// 	fprintf (stderr, "Fontconfig error: No writable cache directories\n");
	//     if (d)
	// 	FcStrFree (d);
	//     FcStrListDone (list);
	//     if (!cacheDir)
	// 	return FcFalse;

	//     FcDirCacheBasenameMD5 (config, dir, cacheBase);
	//     cacheHashed = FcStrBuildFilename (cacheDir, cacheBase, nil);
	//     FcStrFree (cacheDir);
	//     if (!cacheHashed)
	//         return FcFalse;

	//     if (FcDebug () & FC_DBG_CACHE)
	//         printf ("FcDirCacheWriteDir dir \"%s\" file \"%s\"\n",
	// 		dir, cacheHashed);

	//     atomic = FcAtomicCreate ((FcChar8 *)cacheHashed);
	//     if (!atomic)
	// 	goto bail1;

	//     if (!FcAtomicLock (atomic))
	// 	goto bail3;

	//     fd = FcOpen((char *)FcAtomicNewFile (atomic), O_RDWR | O_CREAT | O_BINARY, 0666);
	//     if (fd == -1)
	// 	goto bail4;

	//     /* Temporarily switch magic to MMAP while writing to file */
	//     magic = cache.magic;
	//     if (magic != FC_CACHE_MAGIC_MMAP)
	// 	cache.magic = FC_CACHE_MAGIC_MMAP;

	//     /*
	//      * Write cache contents to file
	//      */
	//     written = write (fd, cache, cache.size);

	//     /* Switch magic back */
	//     if (magic != FC_CACHE_MAGIC_MMAP)
	// 	cache.magic = magic;

	//     if (written != cache.size)
	//     {
	// 	perror ("write cache");
	// 	goto bail5;
	//     }

	//     close(fd);
	//     if (!FcAtomicReplaceOrig(atomic))
	//         goto bail4;

	//     /* If the file is small, update the cache chain entry such that the
	//      * new cache file is not read again.  If it's large, we don't do that
	//      * such that we reload it, using mmap, which is shared across processes.
	//      */
	//     if (cache.size < FC_CACHE_MIN_MMAP && FcStat (cacheHashed, &cacheStat))
	//     {
	// 	lock_cache ();
	// 	if ((skip = FcCacheFindByAddrUnlocked (cache)))
	// 	{
	// 	    skip.cache_dev = cacheStat.st_dev;
	// 	    skip.cache_ino = cacheStat.st_ino;
	// 	    skip.cacheMtime = cacheStat.st_mtime;
	// #ifdef HAVE_STRUCT_STAT_ST_MTIM
	// 	    skip.cacheMtime_nano = cacheStat.st_mtim.tv_nsec;
	// #else
	// 	    skip.cacheMtime_nano = 0;
	// #endif
	// 	}
	// 	unlock_cache ();
	//     }

	//     FcStrFree (cacheHashed);
	//     FcAtomicUnlock (atomic);
	//     FcAtomicDestroy (atomic);
	//     return FcTrue;

	//  bail5:
	//     close (fd);
	//  bail4:
	//     FcAtomicUnlock (atomic);
	//  bail3:
	//     FcAtomicDestroy (atomic);
	//  bail1:
	//     FcStrFree (cacheHashed);
}

// FcBool
// FcDirCacheClean (const FcChar8 *cacheDir, FcBool verbose)
// {
//     DIR		*d;
//     struct dirent *ent;
//     FcChar8	*dir;
//     FcBool	ret = FcTrue;
//     FcBool	remove;
//     FcCache	*cache;
//     struct stat	target_stat;
//     const FcChar8 *sysroot;
//     FcConfig	*config;

//     config = FcConfigReference (nil);
//     if (!config)
// 	return FcFalse;
//     /* FIXME: this API needs to support non-current FcConfig */
//     sysroot = config.getSysRoot ();
//     if (sysroot)
// 	dir = FcStrBuildFilename (sysroot, cacheDir, nil);
//     else
// 	dir = FcStrCopyFilename (cacheDir);
//     if (!dir)
//     {
// 	fprintf (stderr, "Fontconfig error: %s: out of memory\n", cacheDir);
// 	ret = FcFalse;
// 	goto bail;
//     }
//     if (access ((char *) dir, W_OK) != 0)
//     {
// 	if (verbose || FcDebug () & FC_DBG_CACHE)
// 	    printf ("%s: not cleaning %s cache directory\n", dir,
// 		    access ((char *) dir, F_OK) == 0 ? "unwritable" : "non-existent");
// 	goto bail0;
//     }
//     if (verbose || FcDebug () & FC_DBG_CACHE)
// 	printf ("%s: cleaning cache directory\n", dir);
//     d = opendir ((char *) dir);
//     if (!d)
//     {
// 	perror ((char *) dir);
// 	ret = FcFalse;
// 	goto bail0;
//     }
//     while ((ent = readdir (d)))
//     {
// 	FcChar8	*file_name;
// 	const FcChar8	*target_dir;

// 	if (ent.d_name[0] == '.')
// 	    continue;
// 	/* skip cache files for different architectures and */
// 	/* files which are not cache files at all */
// 	if (strlen(ent.d_name) != 32 + strlen ("-" FC_ARCHITECTURE FC_CACHE_SUFFIX) ||
// 	    strcmp(ent.d_name + 32, "-" FC_ARCHITECTURE FC_CACHE_SUFFIX))
// 	    continue;

// 	file_name = FcStrBuildFilename (dir, (FcChar8 *)ent.d_name, nil);
// 	if (!file_name)
// 	{
// 	    fprintf (stderr, "Fontconfig error: %s: allocation failure\n", dir);
// 	    ret = FcFalse;
// 	    break;
// 	}
// 	remove = FcFalse;
// 	cache = FcDirCacheLoadFile (file_name, nil);
// 	if (!cache)
// 	{
// 	    if (verbose || FcDebug () & FC_DBG_CACHE)
// 		printf ("%s: invalid cache file: %s\n", dir, ent.d_name);
// 	    remove = FcTrue;
// 	}
// 	else
// 	{
// 	    FcChar8 *s;

// 	    target_dir = FcCacheDir (cache);
// 	    if (sysroot)
// 		s = FcStrBuildFilename (sysroot, target_dir, nil);
// 	    else
// 		s = FcStrdup (target_dir);
// 	    if (stat ((char *) s, &target_stat) < 0)
// 	    {
// 		if (verbose || FcDebug () & FC_DBG_CACHE)
// 		    printf ("%s: %s: missing directory: %s \n",
// 			    dir, ent.d_name, s);
// 		remove = FcTrue;
// 	    }
// 	    FcDirCacheUnload (cache);
// 	    FcStrFree (s);
// 	}
// 	if (remove)
// 	{
// 	    if (unlink ((char *) file_name) < 0)
// 	    {
// 		perror ((char *) file_name);
// 		ret = FcFalse;
// 	    }
// 	}
//         FcStrFree (file_name);
//     }

//     closedir (d);
// bail0:
//     FcStrFree (dir);
// bail:
//     FcConfigDestroy (config);

//     return ret;
// }

// int
// FcDirCacheLock (dir string,
// 		FcConfig      *config)
// {
//     FcChar8 *cacheHashed = nil;
//     FcChar8 cacheBase[CACHEBASE_LEN];
//     FcStrList *list;
//     FcChar8 *cacheDir;
//     const FcChar8 *sysroot = config.getSysRoot ();
//     int fd = -1;

//     FcDirCacheBasenameMD5 (config, dir, cacheBase);
//     list = FcStrListCreate (config.cacheDirs);
//     if (!list)
// 	return -1;

//     while ((cacheDir = FcStrListNext (list)))
//     {
// 	if (sysroot)
// 	    cacheHashed = FcStrBuildFilename (sysroot, cacheDir, cacheBase, nil);
// 	else
// 	    cacheHashed = FcStrBuildFilename (cacheDir, cacheBase, nil);
// 	if (!cacheHashed)
// 	    break;
// 	fd = FcOpen ((const char *)cacheHashed, O_RDWR);
// 	FcStrFree (cacheHashed);
// 	/* No caches in that directory. simply retry with another one */
// 	if (fd != -1)
// 	{
// #if defined(_WIN32)
// 	    if (_locking (fd, _LK_LOCK, 1) == -1)
// 		goto bail;
// #else
// 	    struct flock fl;

// 	    fl.l_type = F_WRLCK;
// 	    fl.l_whence = SEEK_SET;
// 	    fl.l_start = 0;
// 	    fl.l_len = 0;
// 	    fl.l_pid = getpid ();
// 	    if (fcntl (fd, F_SETLKW, &fl) == -1)
// 		goto bail;
// #endif
// 	    break;
// 	}
//     }
//     FcStrListDone (list);
//     return fd;
// bail:
//     FcStrListDone (list);
//     if (fd != -1)
// 	close (fd);
//     return -1;
// }

// void
// FcDirCacheUnlock (int fd)
// {
//     if (fd != -1)
//     {
// #if defined(_WIN32)
// 	_locking (fd, _LK_UNLCK, 1);
// #else
// 	struct flock fl;

// 	fl.l_type = F_UNLCK;
// 	fl.l_whence = SEEK_SET;
// 	fl.l_start = 0;
// 	fl.l_len = 0;
// 	fl.l_pid = getpid ();
// 	fcntl (fd, F_SETLK, &fl);
// #endif
// 	close (fd);
//     }
// }

// /*
//  * Hokey little macro trick to permit the definitions of C functions
//  * with the same name as CPP macros
//  */
// #define args1(x)	    (x)
// #define args2(x,y)	    (x,y)

// const FcChar8 *
// FcCacheDir args1(const FcCache *c)
// {
//     return FcCacheDir (c);
// }

// FcFontSet *
// FcCacheCopySet args1(const FcCache *c)
// {
//     FcFontSet	*old = FcCacheSet (c);
//     FcFontSet	*new = FcFontSetCreate ();
//     int		i;

//     if (!new)
// 	return nil;
//     for (i = 0; i < old.nfont; i++)
//     {
// 	FcPattern   *font = FcFontSetFont (old, i);

// 	FcPatternReference (font);
// 	if (!FcFontSetAdd (new, font))
// 	{
// 	    FcFontSetDestroy (new);
// 	    return nil;
// 	}
//     }
//     return new;
// }

// const FcChar8 *
// FcCacheSubdir args2(const FcCache *c, int i)
// {
//     return FcCacheSubdir (c, i);
// }

// int
// FcCacheNumSubdir args1(const FcCache *c)
// {
//     return c.dirs_count;
// }

// int
// FcCacheNumFont args1(const FcCache *c)
// {
//     return FcCacheSet(c).nfont;
// }

// FcBool
// FcDirCacheCreateTagFile (const FcChar8 *cacheDir)
// {
//     FcChar8		*cache_tag;
//     int 		 fd;
//     FILE		*fp;
//     FcAtomic		*atomic;
//     static const FcChar8 cache_tag_contents[] =
// 	"Signature: 8a477f597d28d172789f06886806bc55\n"
// 	"# This file is a cache directory tag created by fontconfig.\n"
// 	"# For information about cache directory tags, see:\n"
// 	"#       http://www.brynosaurus.com/cachedir/\n";
//     static size_t	 cache_tag_contents_size = sizeof (cache_tag_contents) - 1;
//     FcBool		 ret = FcFalse;

//     if (!cacheDir)
// 	return FcFalse;

//     if (access ((char *) cacheDir, W_OK) == 0)
//     {
// 	/* Create CACHEDIR.TAG */
// 	cache_tag = FcStrBuildFilename (cacheDir, "CACHEDIR.TAG", nil);
// 	if (!cache_tag)
// 	    return FcFalse;
// 	atomic = FcAtomicCreate ((FcChar8 *)cache_tag);
// 	if (!atomic)
// 	    goto bail1;
// 	if (!FcAtomicLock (atomic))
// 	    goto bail2;
// 	fd = FcOpen((char *)FcAtomicNewFile (atomic), O_RDWR | O_CREAT, 0644);
// 	if (fd == -1)
// 	    goto bail3;
// 	fp = fdopen(fd, "wb");
// 	if (fp == nil)
// 	    goto bail3;

// 	fwrite(cache_tag_contents, cache_tag_contents_size, sizeof (FcChar8), fp);
// 	fclose(fp);

// 	if (!FcAtomicReplaceOrig(atomic))
// 	    goto bail3;

// 	ret = FcTrue;
//       bail3:
// 	FcAtomicUnlock (atomic);
//       bail2:
// 	FcAtomicDestroy (atomic);
//       bail1:
// 	FcStrFree (cache_tag);
//     }

//     if (FcDebug () & FC_DBG_CACHE)
//     {
// 	if (ret)
// 	    printf ("Created CACHEDIR.TAG at %s\n", cacheDir);
// 	else
// 	    printf ("Unable to create CACHEDIR.TAG at %s\n", cacheDir);
//     }

//     return ret;
// }

// void
// FcCacheCreateTagFile (config *FcConfig)
// {
//     FcChar8   *cacheDir = nil, *d = nil;
//     FcStrList *list;
//     const FcChar8 *sysroot;

//     config = FcConfigReference (config);
//     if (!config)
// 	return;
//     sysroot = config.getSysRoot ();

//     list = FcConfigGetCacheDirs (config);
//     if (!list)
// 	goto bail;

//     while ((cacheDir = FcStrListNext (list)))
//     {
// 	if (d)
// 	    FcStrFree (d);
// 	if (sysroot)
// 	    d = FcStrBuildFilename (sysroot, cacheDir, nil);
// 	else
// 	    d = FcStrCopyFilename (cacheDir);
// 	if (FcDirCacheCreateTagFile (d))
// 	    break;
//     }
//     if (d)
// 	FcStrFree (d);
//     FcStrListDone (list);
// bail:
//     FcConfigDestroy (config);
// }
