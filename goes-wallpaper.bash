set -exo pipefail

# save stdout and stderr to file descriptors 3 and 4, then redirect them to "foo"
exec 3>&1 4>&2 >/tmp/last-goes-run.log 2>&1

GOES="https://cdn.star.nesdis.noaa.gov/GOES16/ABI/CONUS/GEOCOLOR/"
curl $GOES$(curl -q $GOES 2>/dev/null \
        | grep GEOCOLOR \
        | cut -d\" -f2 \
        | grep 2500 \
        | tail -n1) \
     2>/dev/null > /tmp/.goes-wallpaper.jpg \
    && DISPLAY=:0 gsettings set \
          org.gnome.desktop.background picture-uri file:///tmp/.goes-wallpaper.jpg

# restore stdout and stderr
exec 1>&3 2>&4
