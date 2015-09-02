dir=$(dirname $0)
NUM_WIN_CELLS=1 $dir/run_wats.sh $dir/bosh_lite_config.json $@
