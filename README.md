# mrun
## Introduction
This repository contains the mrun utility. MRrun executes one or more
instances of the same program in parallel and makes it easy to distinguish
between their outputs.

The original use case is the development of distributed programs: it should be
easy to spawn several copies of the same program, let them connect together
and see how they behave.

## Usage
### Command line
The simplest example executes multiple instances of the `echo` program, each
one printing a message containing its instance identifier:

```
mrun 3 echo "I'm instance {{.InstanceId}}."
```

See `mrun -h` for more information.

### Output
MRun captures the output of the program it executes and print it line by line
on the standard output with a prefix including the identifier of the instance.

Additionally, status messages such as process errors are printed on the error
output, with an extra '#' character.

# Licensing
MRun is open source software distributed under the
[ISC](https://opensource.org/licenses/ISC) license.
