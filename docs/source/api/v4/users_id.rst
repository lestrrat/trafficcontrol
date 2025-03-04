..
..
.. Licensed under the Apache License, Version 2.0 (the "License");
.. you may not use this file except in compliance with the License.
.. You may obtain a copy of the License at
..
..     http://www.apache.org/licenses/LICENSE-2.0
..
.. Unless required by applicable law or agreed to in writing, software
.. distributed under the License is distributed on an "AS IS" BASIS,
.. WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
.. See the License for the specific language governing permissions and
.. limitations under the License.
..

.. _to-api-users-id:

****************
``users/{{ID}}``
****************

``GET``
=======
Retrieves a specific user.

:Auth. Required: Yes
:Roles Required: None
:Permissions Required: USER:READ
:Response Type:  Array

Request Structure
-----------------
.. table:: Request Path Parameters

	+------+-------------------------------------------------------------+
	| Name |                       Description                           |
	+======+=============================================================+
	|  ID  | The integral, unique identifier of the user to be retrieved |
	+------+-------------------------------------------------------------+

.. code-block:: http
	:caption: Request Example

	GET /api/4.0/users/2 HTTP/1.1
	Host: trafficops.infra.ciab.test
	User-Agent: curl/7.47.0
	Accept: */*
	Cookie: mojolicious=...

Response Structure
------------------
:addressLine1:      The user's address - including street name and number
:addressLine2:      An additional address field for e.g. apartment number
:city:              The name of the city wherein the user resides
:company:           The name of the company for which the user works
:country:           The name of the country wherein the user resides
:email:             The user's email address
:fullName:          The user's full name, e.g. "John Quincy Adams"
:gid:               A deprecated field only kept for legacy compatibility reasons that used to contain the UNIX group ID of the user - now it is always ``null``
:id:                An integral, unique identifier for this user
:lastAuthenticated: The date and time at which the user was last authenticated, in :rfc:`3339`
:lastUpdated:       The date and time at which the user was last modified, in :ref:`non-rfc-datetime`
:newUser:           A meta field with no apparent purpose that is usually ``null`` unless explicitly set during creation or modification of a user via some API endpoint
:phoneNumber:       The user's phone number
:postalCode:        The postal code of the area in which the user resides
:publicSshKey:      The user's public key used for the SSH protocol
:registrationSent:  If the user was created using the :ref:`to-api-users-register` endpoint, this will be the date and time at which the registration email was sent - otherwise it will be ``null``
:role:              The name of the role assigned to this user
:stateOrProvince:   The name of the state or province where this user resides
:tenant:            The name of the tenant to which this user belongs
:tenantId:          The integral, unique identifier of the tenant to which this user belongs
:uid:               A deprecated field only kept for legacy compatibility reasons that used to contain the UNIX user ID of the user - now it is always ``null``
:username:          The user's username

.. code-block:: http
	:caption: Response Example

	HTTP/1.1 200 OK
	Access-Control-Allow-Credentials: true
	Access-Control-Allow-Headers: Origin, X-Requested-With, Content-Type, Accept, Set-Cookie, Cookie
	Access-Control-Allow-Methods: POST,GET,OPTIONS,PUT,DELETE
	Access-Control-Allow-Origin: *
	Content-Type: application/json
	Set-Cookie: mojolicious=...; Path=/; Expires=Mon, 18 Nov 2019 17:40:54 GMT; Max-Age=3600; HttpOnly
	Whole-Content-Sha512: 9vqUmt8fWEuDb+9LQJ4sGbbF4Z0a7uNyBNSWhyzAi3fBUZ5mGhd4Jx5IuSlEqiLZnYeViJJL8mpRortkHCgp5Q==
	X-Server-Name: traffic_ops_golang/
	Date: Thu, 13 Dec 2018 17:46:00 GMT
	Content-Length: 454

	{ "response": [
		{
			"addressLine1": null,
			"addressLine2": null,
			"changeLogCount": null,
			"city": null,
			"company": null,
			"country": null,
			"email": null,
			"fullName": null,
			"gid": null,
			"id": 2,
			"lastAuthenticated": "0001-01-01T00:00:00Z",
			"lastUpdated": "2021-08-25T14:08:13.974447-06:00",
			"newUser": false,
			"phoneNumber": null,
			"postalCode": null,
			"publicSshKey": null,
			"registrationSent": null,
			"role": "admin",
			"stateOrProvince": null,
			"tenant": "root",
			"tenantId": 1,
			"uid": null,
			"username": "admin"
		}
	]}

``PUT``
=======

:Auth. Required: Yes
:Roles Required: "admin" or "operations"
:Permissions Required: USER:UPDATE, USER:READ
:Response Type:  Object

Request Structure
-----------------
.. table:: Request Path Parameters

	+------+------------------------------------------------------------+
	| Name |                       Description                          |
	+======+============================================================+
	|  ID  | The integral, unique identifier of the user to be modified |
	+------+------------------------------------------------------------+

:addressLine1:       An optional field which should contain the user's address - including street name and number
:addressLine2:       An optional field which should contain an additional address field for e.g. apartment number
:city:               An optional field which should contain the name of the city wherein the user resides
:company:            An optional field which should contain the name of the company for which the user works
:confirmLocalPasswd: The 'confirm' field in a new user's password specification - must match ``localPasswd``
:country:            An optional field which should contain the name of the country wherein the user resides
:email:              The user's email address The given email is validated (circuitously) by `GitHub user asaskevich's regular expression <https://github.com/asaskevich/govalidator/blob/9a090521c4893a35ca9a228628abf8ba93f63108/patterns.go#L7>`_ . Note that it can't actually distinguish a valid, deliverable, email address but merely ensure the email is in a commonly-found format.
:fullName:           The user's full name, e.g. "John Quincy Adams"
:localPasswd:        The user's password
:newUser:            An optional meta field with no apparent purpose - don't use this
:phoneNumber:        An optional field which should contain the user's phone number
:postalCode:         An optional field which should contain the user's postal code
:publicSshKey:       An optional field which should contain the user's public encryption key used for the SSH protocol
:role:               The number that corresponds to the highest permission role which will be permitted to the user
:stateOrProvince:    An optional field which should contain the name of the state or province in which the user resides
:tenantId:           The integral, unique identifier of the tenant to which the new user shall belong

	.. note:: This field is optional if and only if tenancy is not enabled in Traffic Control

:username: The user's username

.. code-block:: http
	:caption: Request Structure

	PUT /api/4.0/users/2 HTTP/1.1
	Host: trafficops.infra.ciab.test
	User-Agent: curl/7.47.0
	Accept: */*
	Cookie: mojolicious=...
	Content-Length: 458
	Content-Type: application/json

	{
		"addressLine1": "not a real address",
		"addressLine2": "not a real address either",
		"city": "not a real city",
		"company": "not a real company",
		"country": "not a real country",
		"email": "not@real.email",
		"fullName": "Not a real fullName",
		"phoneNumber": "not a real phone number",
		"postalCode": "not a real postal code",
		"publicSshKey": "not a real ssh key",
		"stateOrProvince": "not a real state or province",
		"tenantId": 1,
		"role": "admin",
		"username": "admin"
	}

Response Structure
------------------
:addressLine1:     The user's address - including street name and number
:addressLine2:     An additional address field for e.g. apartment number
:city:             The name of the city wherein the user resides
:company:          The name of the company for which the user works
:country:          The name of the country wherein the user resides
:email:            The user's email address
:fullName:         The user's full name, e.g. "John Quincy Adams"
:gid:              A deprecated field only kept for legacy compatibility reasons that used to contain the UNIX group ID of the user - now it is always ``null``
:id:               An integral, unique identifier for this user
:lastUpdated:      The date and time at which the user was last modified, in :ref:`non-rfc-datetime`
:newUser:          A meta field with no apparent purpose that is usually ``null`` unless explicitly set during creation or modification of a user via some API endpoint
:phoneNumber:      The user's phone number
:postalCode:       The postal code of the area in which the user resides
:publicSshKey:     The user's public key used for the SSH protocol
:registrationSent: If the user was created using the :ref:`to-api-users-register` endpoint, this will be the date and time at which the registration email was sent - otherwise it will be ``null``
:role:             The name of the role assigned to this user
:stateOrProvince:  The name of the state or province where this user resides
:tenant:           The name of the tenant to which this user belongs
:tenantId:         The integral, unique identifier of the tenant to which this user belongs
:uid:              A deprecated field only kept for legacy compatibility reasons that used to contain the UNIX user ID of the user - now it is always ``null``
:username:         The user's username

.. code-block:: http
	:caption: Response Example

	HTTP/1.1 200 OK
	Access-Control-Allow-Credentials: true
	Access-Control-Allow-Headers: Origin, X-Requested-With, Content-Type, Accept
	Access-Control-Allow-Methods: POST,GET,OPTIONS,PUT,DELETE
	Access-Control-Allow-Origin: *
	Cache-Control: no-cache, no-store, max-age=0, must-revalidate
	Content-Type: application/json
	Date: Thu, 13 Dec 2018 17:24:23 GMT
	X-Server-Name: traffic_ops_golang/
	Set-Cookie: mojolicious=...; Path=/; Expires=Mon, 18 Nov 2019 17:40:54 GMT; Max-Age=3600; HttpOnly
	Vary: Accept-Encoding
	Whole-Content-Sha512: QKvGSIwSdreMI/OdgWv9WQfI/C1JbXSoQGGospTGfCVUJ32XNWMhmREGzojWsilW8os8b14TGYeyMLUWunf2Ug==
	Content-Length: 478

	{ "alerts": [
		{
			"level": "success",
			"text": "user was updated."
		}
	],
	"response": {
		"addressLine1": "not a real address",
		"addressLine2": "not a real address either",
		"changeLogCount": null,
		"city": "not a real city",
		"company": "not a real company",
		"country": "not a real country",
		"email": "not@real.email",
		"fullName": "Not a real fullName",
		"gid": null,
		"id": 2,
		"lastAuthenticated": "2021-07-09T14:44:10.371708-06:00",
		"lastUpdated": "2021-08-25T15:05:16.32163-06:00",
		"newUser": false,
		"phoneNumber": "not a real phone number",
		"postalCode": "not a real postal code",
		"publicSshKey": "not a real ssh key",
		"registrationSent": null,
		"role": "admin",
		"stateOrProvince": "not a real state or province",
		"tenant": "root",
		"tenantId": 1,
		"uid": null,
		"username": "admin"
	}}
