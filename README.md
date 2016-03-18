# rumpacker
Rump kernel packing tool

# AWS credentials

Create `~/.aws/credentials` and put following content inside:
```
[default]
aws_access_key_id = AKID1234567890
aws_secret_access_key = MY-SECRET-KEY
```

# Test stuff

- volume id: `vol-96207d49`
- detach
- create snapshot
- create image
- create ami `virtualization_type: paravirtual, kernel_id: aki-919dcaf8 - pv-grub-hd0_1.04-x86_64` - us-east-1 region
- setat attribute AMI-a na public - opcionalno za sad
