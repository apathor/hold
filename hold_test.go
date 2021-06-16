package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBasicUse(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "hold-test")
	assert.NoError(t, err, "temporary directory for testing exists")

	hold, err := NewHoldDir(dir)
	assert.NoError(t, err, "construct new hold directory")
	assert.DirExists(t, hold.path, "the directory exists")
	assert.Equal(t, dir, hold.path, "definitely the expected directory")

	input1 := []byte(`
  Hold \Hold\, v. t. [imp. & p. p. {Held}; p. pr. & vb. n.
     {Holding}. {Holden}, p. p., is obs. in elegant writing,
     though still used in legal language.] [OE. haldan, D. houden,
     OHG. hoten, Icel. halda, Dan. holde, Sw. h[*a]lla, Goth.
     haldan to feed, tend (the cattle); of unknown origin. Gf.
     {Avast}, {Halt}, {Hod}.]
     [1913 Webster]
     1. To cause to remain in a given situation, position, or
        relation, within certain limits, or the like; to prevent
        from falling or escaping; to sustain; to restrain; to keep
        in the grasp; to retain.
        [1913 Webster]

              The loops held one curtain to another. --Ex. xxxvi.
                                                    12.
        [1913 Webster]

              Thy right hand shall hold me.         --Ps. cxxxix.
                                                    10.
        [1913 Webster]

              They all hold swords, being expert in war. --Cant.
                                                    iii. 8.
        [1913 Webster]

              In vain he seeks, that having can not hold.
                                                    --Spenser.
        [1913 Webster]

              France, thou mayst hold a serpent by the tongue, . .
              .
              A fasting tiger safer by the tooth,
              Than keep in peace that hand which thou dost hold.
                                                    --Shak.
        [1913 Webster]`)

	delay := 1 * time.Second

	out1, path1, err1 := hold.Stash("t1", []byte(input1))
	assert.NoError(t, err1, "stash some data")
	assert.Equal(t, out1, input1, "stashed bytes match those returned")
	assert.FileExists(t, path1, "cache file exists")

	out2, path2, err2 := hold.Retrieve("t1", time.Now().Add(-delay))
	assert.NoError(t, err2, "retreive cached file")
	assert.Equal(t, out2, input1, "retreived bytes match input1")
	assert.FileExists(t, path2, "cache file still exists")

	time.Sleep(delay)
	out3, path3, err3 := hold.Retrieve("t1", time.Now())
	assert.Error(t, err3, "cannot retreive expired cache")
	assert.Nil(t, out3, path3, "nothing returned on expried retreive")

	out4, path4, err4 := hold.Retrieve("t2", time.Now())
	assert.Error(t, err4, "cannot retreive unknown cache")
	assert.Nil(t, out4, "output not returned on failed retreive")
	assert.Equal(t, "", path4, "path not returned on failed retreive")

	caches, err := hold.Caches()
	assert.NoError(t, err2, "caches can be listed")
	assert.Equal(t, "t1", caches[0], "cache names are returned")
}
