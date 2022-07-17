import logging
import os

import grpc
from google.protobuf.json_format import Parse, ParseDict

import defn.dev.legumes.v1.bean_pb2 as bean_pb2
import defn.dev.legumes.v1.bean_pb2_grpc as bean_pb2_grpc


os.environ["GRPC_VERBOSITY"] = "DEBUG"

with open(os.environ["GRPC_DEFAULT_SSL_ROOTS_FILE_PATH"], 'rb') as f:
    certificate_chain = f.read()

def run():
    with grpc.secure_channel(
        os.environ.get("server", "kourier-internal-x-kourier-system-x-vc1.vc1.svc:80"),
        grpc.ssl_channel_credentials(root_certificates=certificate_chain),
        options=[
            (
                "grpc.default_authority",
                os.environ.get("authority", "hello.demo.svc.cluster.local"),
            )
        ],
    ) as channel:
        stub = bean_pb2_grpc.BeanStoreServiceStub(channel)

        response: bean_pb2.Bean = stub.GetBean(
            bean_pb2.Bean(
                url=os.environ.get("url", "cool"),
                sha256=os.environ.get("sha256", "beans"),
            )
        )
        print(f"Bean client received: {response.url}, {response.sha256}")


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    run()