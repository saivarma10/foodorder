from flask import Flask, jsonify
from flask import request

app = Flask(__name__)

@app.route('/hello')
def hello_world():
    return jsonify(message="Hello from Flask API!")

@app.route('/deliveryinitiated', methods=['POST'])
def delivery_initiated():
    
    # Get delivery details from request
    data = request.get_json()
    
    order_id = data.get('orderId')
    customer_name = data.get('customerName')
    delivery_address = data.get('deliveryAddress')
    phone_number = data.get('phoneNumber')
    items = data.get('items', [])
    total_amount = data.get('totalAmount')
    
    # Prepare response with delivery details
    response = {
        'status': 'Delivery Initiated',
        'message': 'Your delivery has been initiated successfully!',
        'deliveryDetails': {
            'orderId': order_id,
            'customerName': customer_name,
            'deliveryAddress': delivery_address,
            'phoneNumber': phone_number,
            'items': items,
            'totalAmount': total_amount
        }
    }
    
    return jsonify(response)

if __name__ == '__main__':
    app.run(host='0.0.0.0', debug=True, port=8010)