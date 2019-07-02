#!/usr/bin/python3

from __future__ import absolute_import
from __future__ import print_function
from __future__ import unicode_literals

import pexpect
import sys

# Note that, for Python 3 compatibility reasons, we are using spawnu and
# importing unicode_literals (above). spawnu accepts Unicode input and
# unicode_literals makes all string literals in this script Unicode by default.
child = pexpect.spawnu('./run.sh')

child.expect('Enter a passphrase to encrypt your key to disk:')
child.sendline('12345678')
child.expect('Repeat the passphrase:')
child.sendline('12345678')
print('>>>>>>1')
child.expect('Enter a passphrase to encrypt your key to disk:')
child.sendline('12345678')
child.expect('Repeat the passphrase:')
child.sendline('12345678')
print('>>>>>>2')
child.expect('Enter a passphrase to encrypt your key to disk:')
child.sendline('12345678')
child.expect('Repeat the passphrase:')
child.sendline('12345678')
print('>>>>>>3')
child.expect('Enter a passphrase to encrypt your key to disk:')
child.sendline('12345678')
child.expect('Repeat the passphrase:')
child.sendline('12345678')
print('>>>>>>4')
child.interact()
#child.expect('Password to sign with validator0:')
#child.sendline('12345678')
#print('>>>>>done')

