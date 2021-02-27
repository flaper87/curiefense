"""
this script converts BQ JSON structure to ES structure
 input stdin
output stdout
"""

import fileinput
import json
import sys
import datetime
import argparse
import socket


parser = argparse.ArgumentParser()

parser.add_argument(
    "--spoof", default=False, action="store_true", help="spoof private data"
)

from metrics import process_metrics

atbash = {
    "a": "z",
    "A": "Z",
    "b": "y",
    "B": "Y",
    "c": "x",
    "C": "X",
    "d": "w",
    "D": "W",
    "e": "v",
    "E": "V",
    "f": "u",
    "F": "U",
    "g": "t",
    "G": "T",
    "h": "s",
    "H": "S",
    "i": "r",
    "I": "R",
    "j": "q",
    "J": "Q",
    "k": "p",
    "K": "P",
    "l": "o",
    "L": "O",
    "M": "N",
    "m": "n",
    "N": "M",
    "n": "m",
    "o": "l",
    "O": "L",
    "p": "k",
    "P": "K",
    "q": "j",
    "Q": "J",
    "r": "i",
    "R": "I",
    "s": "h",
    "S": "H",
    "t": "g",
    "T": "G",
    "u": "f",
    "U": "F",
    "v": "e",
    "V": "E",
    "w": "d",
    "W": "D",
    "x": "c",
    "X": "C",
    "y": "b",
    "Y": "B",
    "z": "a",
    "Z": "A",
}
atbash["0"] = "9"
atbash["1"] = "8"
atbash["2"] = "7"
atbash["3"] = "6"
atbash["4"] = "5"
atbash["5"] = "4"
atbash["6"] = "3"
atbash["7"] = "2"
atbash["8"] = "1"
atbash["9"] = "0"


def spoof_text(_input):
    try:
        output = (atbash.get(c, c) for c in _input)
        return "".join(output).encode("rot13")
    except:
        return _input


def sppof_ip(ipaddr):
    try:
        ipnum = struct.unpack("!I", socket.inet_aton(ipaddr))[0]
        ipnum = ipnum ^ 24
        return socket.inet_ntoa(struct.pack("!I", ipnum))
    except:
        return ipaddr


def spoof_record(record):
    record["authority"] = sppof_text(record["authority"])
    record["path"] = sppof_text(record["path"])

    spoofed_client_ip = spoof_ip(record["downstream"]["remoteaddress"])
    spoofed_origin_ip = spoof_ip(record["upstream"]["remoteaddress"])

    record["downstream"]["remoteaddress"] = spoofed_client_ip
    record["upstream"]["remoteaddress"] = spoofed_origin_ip
    # record["headers"]["host"] = sppof_text(record["headers"]["host"])

    record["request"]["originalpath"] = sppof_text(record["request"]["originalpath"])
    # xff = request["headers"].get("x-forwarded-for")
    # if xff:
    #     request["headers"]["x-forwarded-for"] = ", ".join((sppof_ip(i) for i in xff.split(", ")))

    for n, v in request.get("cookies", {}).items():
        request["cookies"][n] = spoof_text(v)

    for n, v in request.get("arguments", {}).items():
        request["arguments"][n] = spoof_text(v)

    for n, v in request.get("headers", {}).items():
        request["headers"][n] = spoof_text(v)

    tags = [t for t in record["tags"] if not t.startswith("ip-")]
    tags.append("ip-%s" % spoofed_client_ip.replace(".", "-"))
    record["tags"] = tags

    return record


def dictify_headers(request):
    block = request["request_headers"]
    ret = dict((i["h_name"], i["h_value"]) for i in block)
    ret["referer"] = request["referer"]
    ret["user-agent"] = request["user_agent"]
    ret["host"] = request["host"]
    return ret


def dictify_cookies(block):
    return dict((i["c_name"], i["c_value"]) for i in block)


def dictify_args(block):
    return dict((i["arg_name"], i["arg_value"]) for i in block)


def tagify(tags, record):
    tags = list(tags)

    if record["is_anonymizer"]:
        tags.append("anonymizer")

    if record["is_human"]:
        tags.append("human")
    else:
        tags.append("bot")

    if record["is_proxy"]:
        tags.append("proxy")

    if record["is_tor"]:
        tags.append("tor")

    if record["is_vpn"]:
        tags.append("vpn")

    return tags


def reblazer_to_elastic(ngxjson, timestamp):
    anything_else = json.loads(ngxjson.get("anything_else", "{}"))

    eljson = {}

    eljson["requestid"] = ngxjson["request_id"]
    eljson["@timestamp"] = timestamp

    tls_used = len(ngxjson["ssl_protocol"]) > 2
    eljson["scheme"] = (tls_used and "https") or "http"
    eljson["authority"] = ngxjson["host"]
    eljson["port"] = ngxjson["port"] = (tls_used and 443) or 80
    eljson["method"] = ngxjson["method"]
    eljson["path"] = ngxjson["uri"]
    eljson["blocked"] = ngxjson["blocked"]
    eljson["block_reason"] = {"reason": ngxjson["block_reason"]}

    eljson["downstream"] = {"remoteaddress": ngxjson["remote_addr"]}

    if len(ngxjson["upstream_addr"]) > 1:
        upstream = ngxjson["upstream_addr"].split(":")
        addr = upstream[0]
        # port = int(upstream[1].split(",")[0])
        eljson["upstream"] = {"remoteaddress": addr, "remoteaddressport": 0}

    eljson["tls"] = {
        "ciphersuite": ngxjson["ssl_cipher"],
        "version": ngxjson["ssl_protocol"],
    }

    eljson["request"] = {
        "bodybytes": int(ngxjson["request_length"]),
        "originalpath": ngxjson["request"],
        "headers": anything_else["headers"],
        "cookies": anything_else["cookies"],
        "arguments": anything_else["args"],
        # "attributes": dictify_attrs(ngxjson),
    }

    eljson["request"]["headers"]["referer"] = ngxjson["referer"]
    eljson["request"]["headers"]["user-agent"] = ngxjson["user_agent"]
    eljson["request"]["headers"]["host"] = ngxjson["host"]

    eljson["tags"] = tagify(anything_else.get("tags", {}).keys(), ngxjson)

    eljson["response"] = {
        "bodybytes": int(ngxjson["bytes_sent"]),
        "code": int(ngxjson["status"]),
    }

    # return spoof_record(eljson)
    return eljson


def gen_metric_index(recid):
    return {"index": {"_index": "curiemetrics", "_id": recid}}


def gen_accesslog_index(recid):
    return {"index": {"_index": "curieaccesslog", "_id": recid}}


def main():

    for line in fileinput.input():
        try:
            ngxjson = json.loads(line)
            timestamp = datetime.datetime.fromtimestamp(
                float(ngxjson["timestamp"])
            ).isoformat()
            eljson = reblazer_to_elastic(ngxjson, timestamp)
            recid = eljson["requestid"]
            index_line = json.dumps(gen_accesslog_index(recid))
            print(index_line, file=sys.stdout)
            print(json.dumps(eljson), file=sys.stdout)

            metric_doc = process_metrics(eljson, timestamp)
            if metric_doc:
                index_line = json.dumps(gen_metric_index(recid))
                print(index_line, file=sys.stdout)
                print(metric_doc, file=sys.stdout)
        except TypeError:
            raise

        except AttributeError:
            raise

        except Exception as err:
            print(err, file=sys.stderr)
            print("Unexpected error:", sys.exc_info()[0], file=sys.stderr)


if __name__ == "__main__":
    main()
