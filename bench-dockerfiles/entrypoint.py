import time
import argparse

parser = argparse.ArgumentParser('entrypoint script')
parser.add_argument('-o', '--output', type=str, default='output.json')
parser.add_argument('-dn', '--dataset-name', type=str, default='dataset')
parser.add_argument('-d', '--dataset-path', type=str, default='dataset/')
parser.add_argument('-m', '--model-path', type=str, default='sam_checkpoint.pt')

args = parser.parse_args()
print(args)

# simulate running a model benchmark
time.sleep(60)

with open(args.output, 'w') as fs:
    print(args.output)
    fs.write('{"mae": 92.0, "loss": 0.0023}\n')
