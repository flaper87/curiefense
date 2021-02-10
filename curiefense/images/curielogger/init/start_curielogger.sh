#!/bin/bash

ES_INDEX_NAME=${CURIEFENSE_ES_INDEX_NAME:-curieaccesslog}
ES_PATTERN_PATH="api/saved_objects/index-pattern/${ES_INDEX_NAME}"

CURL="curl --silent --write-out %{http_code}\n -H Content-Type:application/json --output /dev/null"

wait_for_es () {
	if $CURL -X GET "${ELASTICSEARCH_URL}_cluster/health?wait_for_status=yellow&timeout=10s"|grep -qv 200; then
		sleep 5
		wait_for_es
	fi
}

create_kibana_index_pattern () {
	# Wait for kibana to become reachable
	while true; do
		if $CURL "${CURIELOGGER_KIBANA_URL}api/status"|grep -q 200; then
			break
		fi
		echo "Kibana at $CURIELOGGER_KIBANA_URL is not reachable yet, waiting 5s..."
		sleep 5
	done

	# Check whether the index pattern already exists
	if $CURL "$CURIELOGGER_KIBANA_URL$ES_PATTERN_PATH"|grep -q 200; then
		# already exists
		echo "Kibana index already exists."
	else
		# Create the index pattern
		if $CURL -X POST "$CURIELOGGER_KIBANA_URL$ES_PATTERN_PATH" -H 'kbn-xsrf: true' -d '{"attributes": {"title": "'"$ES_INDEX_NAME*"'","timeFieldName": "@timestamp"}}'|grep -q 200; then

			echo "Kibana index $ES_INDEX_NAME created."
		else
			echo "Kibana index $ES_INDEX_NAME creation failed, retrying."
			sleep 5
			create_kibana_index_pattern
		fi
	fi
}


if [ -n "$CURIELOGGER_INITIALIZE_KIBANA_INDEX" ]; then
	# in case logs are saved in elasticsearch and not postgres
	>&2 echo "Creating an index pattern in Kibana if needed."
	wait_for_es
	create_kibana_index_pattern
fi

/bin/curielogger
