import argparse
import commons_pb2_grpc
import commons_pb2
import grpc
import random
import base64
import array
import pytz
from datetime import datetime, timedelta

from Crypto.Cipher import AES
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.kdf.hkdf import HKDF
from cryptography.hazmat.backends import default_backend

URL = 'localhost:5000'
PROFESSIONAL_API_KEY='c3b9b61b687b895aff09eb072fb07d33'
backend = default_backend()


class Session:
    def __init__(self, entities=10):
        self.entities = []
        self.time = datetime(2020, 5, 1, tzinfo=pytz.utc)  # May 1st
        for i in range(entities):
            self.entities.append(Entity(f"entity-{i}", self.time))

    def step(self):
        new_time = self.time + timedelta(minutes=15)
        do_test = new_time.day != self.time.day
        if do_test:
            print(new_time)
        step_rpis = []
        for ent in self.entities:
            rpi = ent.step(new_time)
            step_rpis.append(rpi)
            if do_test:
                ent.maybe_test()
        # distribute a random sample of 20% of the RPIs
        exchanged_rpis = random.sample(step_rpis, int(.5 * len(self.entities)))
        for rpi in exchanged_rpis:
            ent = random.choice(self.entities)
            ent.observe(rpi)

        self.time = new_time


class Entity:
    def __init__(self, name, start_timestamp):
        self._name = name
        tek, enin = generate_random_tek()
        self._enins = [enin]
        self._teks = [tek]
        self._time = start_timestamp
        self._rpis = {}
        self._diagnosed = False
        self._tested = False
        self._exposed = False
        self._seen_rpis = []

        channel = grpc.insecure_channel(URL)
        self.stub = commons_pb2_grpc.DiagnosisDBStub(channel)

    def maybe_test(self):
        if random.random() < 0.10:
            self._tested = True
        if self._tested:
            self._diagnosed = random.random() < 0.25
        if self._diagnosed:
            # if they are diagnosed, first need to get the authorization key
            # from the authorized professional

            ############ PROFESSIONAL DOES THIS ############
            print(self._time)
            resp = self.stub.GetAuthorizationToken(commons_pb2.TokenRequest(
                # this is secret
                api_key=bytes.fromhex(PROFESSIONAL_API_KEY),
                permitted_range_start=self._time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                permitted_range_end=(self._time + timedelta(days=14)).strftime("%Y-%m-%dT%H:%M:%SZ"),
                key_type=commons_pb2.DIAGNOSED,
            ))
            if len(resp.error) > 0:
                raise Exception(f"could not get authorization key {resp.error}")
            # then auth_key is given to the user
            ############# BACK TO THE USER NOW ############

            resp = self.stub.AddReport(commons_pb2.Report(
                authorization_key=resp.authorization_key,
                reports=[
                    commons_pb2.TimestampedTEK(
                        TEK=self._teks[-1],
                        ENIN=self._enins[-1]
                    )
                ]
            ))
            if len(resp.error) > 0:
                raise Exception(f"could not upload report: {resp.error}")

    def observe(self, rpi):
        self._seen_rpis.append(rpi)

    def determine_exposure(self):
        enin = dt_to_enin(self._time)
        for r in self.stub.GetDiagnosisKeys(commons_pb2.GetKeyRequest(
            hrange=commons_pb2.HistoricalRange(days=14)
        )):
            if len(r.error) > 0:
                raise Exception(f"could not fetch keys: {r.error}")
            enin = r.record.ENIN
            tek = r.record.TEK
            for i in range(96):
                rpi = compute_rpi(tek, enin+600*i)
                if rpi in self._seen_rpis:
                    when = datetime.utcfromtimestamp(enin * 600)
                    print(f"{self.name} was exposed at {when}")
                    self._exposed = True
                    break

    @property
    def name(self):
        return self._name

    def step(self, timestamp):
        # generate a new TEK for a new day
        if self._time.day != timestamp.day:
            tek = bytes([random.getrandbits(8) for i in range(16)])
            enin = dt_to_enin(timestamp)
            self._teks.append(tek)
            self._enins.append(enin)
            self.determine_exposure()
        self._time = timestamp
        tek = self._teks[-1]
        tekb64 = encodeb64(tek)
        if tekb64 not in self._rpis:
            self._rpis[tekb64] = []
        # generate the RPI for this timestamp
        rpi = compute_rpi(tek, dt_to_enin(timestamp))
        self._rpis[tekb64].append(rpi)
        return rpi


def compute_rpi(tek, enin):
    info = "EN-RPIK".encode("utf8")
    _rpik = HKDF(algorithm=hashes.SHA256(), length=16, info=info,
                 backend=backend, salt=None)
    rpik = _rpik.derive(tek)
    padded_data = array.array('b')
    padded_data.frombytes('EN-RPI'.encode('utf8'))
    for i in range(6):
        padded_data.append(0)
    padded_data.frombytes(enin.to_bytes(4, byteorder='little'))
    padded_data = padded_data.tobytes()
    cipher = AES.new(rpik, AES.MODE_ECB)
    rpi = cipher.encrypt(padded_data)
    return rpi


def dt_to_enin(ts, window_minutes=10):
    return int(datetime.timestamp(ts) / (60 * window_minutes))


def now_to_enin():
    return dt_to_enin(datetime.now())


def generate_random_tek():
    randbytes = [random.getrandbits(8) for i in range(16)]
    return bytes(randbytes), now_to_enin()


def encodeb64(byts):
    return base64.encodebytes(byts).decode('utf8').strip()


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Simulate en-db19 workflow')
    parser.add_argument('-e', '--entities', metavar='entities', type=int, default=10)
    parser.add_argument('-d', '--days', metavar='days', type=int, default=30)
    args = parser.parse_args()
    print(f"Running simulation for {args.entities} entities over {args.days} days")
    s = Session(args.entities)
    for i in range(96*args.days):
        s.step()
    # for e in s.entities:
    #     e.determine_exposure()
