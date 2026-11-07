import sys
sys.dont_write_bytecode=True
from run import *
assert run('head2-apply',['bash','stage3/apply.sh','--build-product',S/'head2'],extra={'STAGE3_PRODUCT_STORE':str(S/'products-head2')})==0
