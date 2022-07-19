import logging
import os
import timeit

import grpc
from google.protobuf.json_format import Parse, ParseDict

import defn.dev.legumes.v1.bean_pb2 as bean_pb2
import defn.dev.legumes.v1.bean_pb2_grpc as bean_pb2_grpc


os.environ["GRPC_VERBOSITY"] = "DEBUG"
os.environ["GRPC_DEFAULT_SSL_ROOTS_FILE_PATH"] = "meh.ca"

with open(os.environ["GRPC_DEFAULT_SSL_ROOTS_FILE_PATH"], "rb") as f:
    certificate_chain = f.read()


def run():
    with grpc.secure_channel(
        os.environ.get("server", "169.254.32.1:443"),
        grpc.ssl_channel_credentials(root_certificates=certificate_chain),
        options=[
            (
                "grpc.default_authority",
                os.environ.get("authority", "hello.demo.svc.cluster.local"),
            )
        ],
    ) as channel:
        for a in range(1, 10):
            print(timeit.timeit(lambda: req(channel), number=1))


def req(ch):
    stub = bean_pb2_grpc.BeanStoreServiceStub(ch)

    for a in range(1, 100):
        response: bean_pb2.Bean = stub.GetBean(
            bean_pb2.Bean(
                url=os.environ.get("url", "cool"),
                sha256=os.environ.get("sha256", "beans"),
            )
        )

    return response


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    run()
