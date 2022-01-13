#!/bin/bash
set -e

SCRIPTS_DIR=${SCRIPTS_DIR:-"$PWD"}
echo "Creating plot..."
if [ -z "$REPORT_FILE" ]; then
  echo "Enter the path to the report file:"
  read REPORT_FILE
fi

cp $SCRIPTS_DIR/plot.gnu.template plot.gnu.tmp
jq '.results |=sort_by(.run_count)|.results[]|.client_duration_ns' $REPORT_FILE > plot.tsv
cat -n plot.tsv > plot.tsv.1 && mv plot.tsv.1 plot.tsv
echo "'plot.tsv' using 1:2 smooth bezier title 'round trip ns' with lines" >> plot.gnu.tmp

# plot the data
gnuplot plot.gnu.tmp

# clean up
rm *.tsv
rm plot.gnu.tmp