from stucco import ContrastSetLearner
from seaborn import load_dataset

frame = load_dataset('diamonds')
learner = ContrastSetLearner(frame, group_feature='color')

learner.learn(max_length=3)

output = learner.score(min_lift=2)

print(output)
