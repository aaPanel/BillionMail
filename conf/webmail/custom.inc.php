<?php
    # Product name
    $config['product_name'] = 'Billion-Mail';

    $config['mime_types'] = '/var/roundcube/config/mime.types';
       
    # Plugins
    $config['plugins'] = array('password','userinfo','newmail_notifier','emoticons','zipdownload');

    # Password
    $config['password_query'] = 'update mailbox set password = %P where username = %u';
    $config['password_algorithm'] = 'md5-crypt';


# Please create an extra file: extra.php. Used to persist overrides Configuration
