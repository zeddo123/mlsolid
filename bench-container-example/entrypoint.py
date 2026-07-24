import time
import argparse
import os
import json
import random

parser = argparse.ArgumentParser('entrypoint script')
parser.add_argument('-o', '--output', type=str, default='output.json')
parser.add_argument('-dn', '--dataset-name', type=str, default='dataset')
parser.add_argument('-d', '--dataset-path', type=str, default='dataset/')
parser.add_argument('-m', '--model-path', type=str, default='sam_checkpoint.pt')

args = parser.parse_args()
print(args)

# simulate running a model benchmark
time.sleep(20)

print('Dataset path:', os.listdir(args.dataset_path))
if os.path.exists(args.model_path):
    print('Model path:', os.stat(args.model_path))
else:
    print('Model path: does not exist')

with open(args.output, 'w') as fs:
    print(args.output)
    metrics = {"mae": random.uniform(0.0, 100.0), "loss": random.uniform(0.0, 1.0)}
    json.dump(metrics, fs)
    fs.write('\n')
