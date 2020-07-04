# ring

RingChannel uses two channels input and output in a way that never blocks the input.
If a new value gets added to a RingChannel's input when its buffer is full then the oldest
value in the output channel gets removed to make room (just like a standard ring-buffer).