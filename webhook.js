class Script {
    prepare_outgoing_request({ request }) {
        const trigger = request.data.trigger_word.toLowerCase() + ' ';
        const phrase = request.data.text.toLowerCase().replace(trigger, '');
        return {
            url: request.url + '?zip=' + phrase,
            method: 'GET'
        };
    }

    process_outgoing_response({ request, response }) {
        return {
            content: {
                text: response.content_raw
            }
        };
    }
}
