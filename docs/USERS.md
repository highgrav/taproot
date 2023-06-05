# Users

While Taproot is agnostic about where your users come from, it does expect a certain "shape" or data. Generally, you
are likely to have to coerce your "native" user format to Taproot's expectations. The main things Taproot expects are:

- A Realm ID (string). The Realm represents where users come from. Most sites only have one Realm, but if you have a "marketplace"-style
application, you could have multiple Realms.
- A Domain ID (string). Domains are subsidiary to Realms. For a single-tenant app, you are likely to have only one Domain, 
but in a multi-tenant app you would have a unique Domain for each tenant. Domains separate out workgroups and labels, as 
described below, so a user can have different workgroups and labels per-domain.
- Workgroups: A Workgroup consists of a Workgroup ID and a Workgroup Name, and Acacia can test either ID or Name. Workgroups
could map to a number of different security concepts beyond Taproot, but essentially all that's important here is that multiple 
users can share Workgroups, and that Workgroup IDs should be (but do not have to be) unique across Domains but Workgroup Names 
need only be unique within a Domain. For more complex scenarios (such as a multi-tenancy environment where a user could have many 
workgroups spread across multiple tenants), `CheckUserRight()` and related functions are useful.
- Rights: `CheckUserRight(userId, domainId, userRight, itemId string)` is the basic function for determining if a user has 
a specific `userRight` within a tenant or domain `userDomain`, optionally on an object with ID `itemId`. User rights are 
assumed to be globally-addressable via a string; for example, you might ask if a user has the `mailbox::email::read` right 
on mailbox ID `12345`. No particular style is enforced other than the right being a unique string.
- Labels: Labels are simple text strings attached to a User within a Domain. Labels are meant to be more informal than Workgroups, 
and thus do not have IDs, just names. For example, you might have a user in the `content::editor` workgroup who has an 
"nyc-ny-us" label that represents their home office. Labels are more freeform and, while they can be tested in Acacia or
as part of business logic, they should be considered much less formal.