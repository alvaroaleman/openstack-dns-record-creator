# Openstack dns record creator

A simple daemon that creates or deletes a DNS entry in CoreDNS whenever
a floating ip gets (de-)associated in your Openstack cloud.

## Configuration

## Configure neutron to publish events

```
crudini --set /etc/nova/neutron.conf oslo_messaging_notifications driver messaging
crudini --set /etc/nova/neutron.conf oslo_messaging_notifications topics notifications
openstack-service restart neutron
```
