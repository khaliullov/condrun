# Conditional Runner

Run different programs depending on arguments passed

## Synopsis

Sample `/etc/default/condrun.yaml`:

    restart-mtproxy:
        command:
            - sudo systemctl restart MTProxy.service
        when:
            arg1:
              - test
            arg2:
              - test2
